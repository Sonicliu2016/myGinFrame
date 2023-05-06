package service

import (
	"errors"
	"fmt"
	"io"
	"math"
	"myGinFrame/glog"
	"myGinFrame/model"
	"myGinFrame/mongodb"
	"myGinFrame/mysql"
	"myGinFrame/redis"
	"myGinFrame/tool"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type FileServie interface {
	//InitBlockUpload : 初始化分块上传
	InitBlockUpload(username, fileHash string, fileSize int) (interface{}, error)
	//UploadBlock : 上传文件分块
	UploadBlock(uploadId, blockHash, blockIndex string, body io.ReadCloser) error
	//CompleteUpload : 通知上传合并
	CompleteUpload(uploadId, fileHash string) error
}

type fileService struct {
	userMysqlDao mysql.UserDao
	userMongoDao mongodb.BaseDao
}

func NewFileService() FileServie {
	return &fileService{
		userMysqlDao: mysql.NewUserDaoManage(),
		userMongoDao: mongodb.NewBaseDaoManage("numbers"),
	}
}

const (
	// ChunkDir : 上传的分块所在目录
	BlockDir = "/file/blocks/"
	// MergeDir : 合并后的文件所在目录
	MergeDir = "/file/merge/"
	// ChunkKeyPrefix : 分块信息对应的redis键前缀
	FileKeyPrefix    = "file_"
	BlockIndexPrefix = "blockIndex_"
	// HashUpIDKeyPrefix : 文件hash映射uploadid对应的redis键前缀
	UploadIdKeyPrefix = "upload_id_"
)

// InitBlockUpload : 初始化分块上传
func (s *fileService) InitBlockUpload(username, fileHash string, fileSize int) (interface{}, error) {
	var uploadId string
	var err error
	// 1. 通过文件hash判断是否断点续传，并获取uploadID
	if redis.CheckKey(UploadIdKeyPrefix + fileHash) {
		uploadId, err = redis.GetString(UploadIdKeyPrefix + fileHash)
		if err != nil {
			return nil, err
		}
	}

	// 2.1 首次上传则新建uploadId
	// 2.2 断点续传则根据uploadId获取已上传的文件分块index
	completedBlockIdxs := []int{}
	if uploadId == "" {
		uploadId = username + fmt.Sprintf("%x", time.Now().UnixNano())
	} else {
		blocks, err := redis.GetHashAll(FileKeyPrefix + uploadId)
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(blocks); i += 2 {
			k := string(blocks[i].([]byte))
			v := string(blocks[i+1].([]byte))
			glog.Glog.Info("k:", k, "->v:", v)
			if strings.HasPrefix(k, BlockIndexPrefix) && v == "ok" {
				// blockIndex_6 -> 6
				blockIdx, _ := strconv.Atoi(k[len(BlockIndexPrefix)+1:])
				completedBlockIdxs = append(completedBlockIdxs, blockIdx)
			}
		}
	}

	// 3. 生成分块上传的初始化信息
	upInfo := &model.BlockFileUploadInfo{
		FileHash:           fileHash,
		FileSize:           fileSize,
		UploadId:           uploadId,
		BlockSize:          5 * 1024 * 1024,                                       // 5MB
		BlockCount:         int(math.Ceil(float64(fileSize) / (5 * 1024 * 1024))), //向上取整
		CompletedBlockIdxs: completedBlockIdxs,
	}

	// 4. 将初始化信息写入到redis缓存
	if len(upInfo.CompletedBlockIdxs) <= 0 {
		hkey := FileKeyPrefix + upInfo.UploadId
		redis.SetHash(hkey, "fileHash", upInfo.FileHash)
		redis.SetHash(hkey, "fileSize", upInfo.FileSize)
		redis.SetHash(hkey, "blockCount", upInfo.BlockCount)
		redis.SetKeyExpire(hkey, 60*60*12)
		redis.SetKeyAndExpire(UploadIdKeyPrefix+fileHash, upInfo.UploadId, 60*60*12)
	}
	return upInfo, nil
}

// UploadBlock : 上传文件分块
func (s *fileService) UploadBlock(uploadId, blockHash, blockIndex string, body io.ReadCloser) error {
	// 1. 获得文件句柄，用于存储分块内容
	fpath := tool.GetConfigStr("staticPath") + BlockDir + uploadId + "/" + blockIndex
	os.MkdirAll(path.Dir(fpath), os.ModePerm)
	fd, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer fd.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, err := body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}

	// 2. 校验分块hash
	cmpSha1, err := tool.ComputeSha1ByShell(fpath)
	if err != nil || cmpSha1 != blockHash {
		glog.Glog.Error("Verify chunk sha1 failed, compare OK: %t, err:%+v\n", cmpSha1 == blockHash, err)
		return err
	}

	// 3. 更新每个块上传完毕的信息
	return redis.SetHash(FileKeyPrefix+uploadId, BlockIndexPrefix+blockIndex, "ok")
}

// CompleteUpload : 通知上传合并
func (s *fileService) CompleteUpload(uploadId, fileHash string) error {
	// 1. 通过uploadid查询redis并判断是否所有分块上传完成
	blocks, err := redis.GetHashAll(FileKeyPrefix + uploadId)
	if err != nil {
		return err
	}
	totalCount := 0
	chunkCount := 0
	for i := 0; i < len(blocks); i += 2 {
		k := string(blocks[i].([]byte))
		v := string(blocks[i+1].([]byte))
		if k == "blockCount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, BlockIndexPrefix) && v == "ok" {
			chunkCount++
		}
	}
	if totalCount != chunkCount {
		glog.Glog.Info("totalCount:", totalCount, "->chunkCount:", chunkCount)
		return errors.New("totalCount != chunkCount")
	}

	// 4. 合并分块 (备注: 更新于2020/04/01; 此合并逻辑非必须实现，因后期转移到ceph/oss时也可以通过分块方式上传)
	if mergeSuc := tool.MergeChuncksByShell(BlockDir+uploadId, MergeDir+fileHash, fileHash); !mergeSuc {
		return errors.New("complete upload failed")
	}

	// 5. 更新唯一文件表及用户文件表
	//fsize, _ := strconv.Atoi(fileSize)
	// 更新于2020-04: 增加fileaddr参数的写入
	//dblayer.OnFileUploadFinished(filehash, filename, int64(fsize), MergeDir+filehash)
	//dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(fsize))

	// 更新于2020-04: 删除已上传的分块文件及redis分块信息
	/*delHashErr := redis.DelKey(UploadIdKeyPrefix + fileHash)
	delChunkErr := redis.DelKey(FileKeyPrefix + uploadId)
	if delChunkErr != nil || delHashErr != nil {
		return errors.New("complete upload part failed")
	}

	delRes := tool.RemovePathByShell(BlockDir + uploadId)
	if !delRes {
		fmt.Printf("Failed to delete chuncks as upload comoleted, uploadID: %s\n", uploadId)
	}*/
	return nil
}

// CancelUploadHandler : 文件取消上传接口
//func CancelUploadHandler(w http.ResponseWriter, r *http.Request) {
//	// 1. 解析用户请求参数
//	r.ParseForm()
//	filehash := r.Form.Get("filehash")
//
//	// 2. 获得redis的一个连接
//	rConn := rPool.RedisPool().Get()
//	defer rConn.Close()
//
//	// 3. 检查uploadID是否存在，如果存在则删除
//	uploadID, err := redis.String(rConn.Do("GET", HashUpIDKeyPrefix+filehash))
//	if err != nil || uploadID == "" {
//		w.Write(util.NewRespMsg(-1, "Cancel upload part failed", nil).JSONBytes())
//		return
//	}
//
//	_, delHashErr := rConn.Do("DEL", HashUpIDKeyPrefix+filehash)
//	_, delUploadInfoErr := rConn.Do("DEL", ChunkKeyPrefix+uploadID)
//	if delHashErr != nil || delUploadInfoErr != nil {
//		w.Write(util.NewRespMsg(-2, "Cancel upload part failed", nil).JSONBytes())
//		return
//	}
//
//	// 4. 删除已上传的分块文件
//	delChkRes := util.RemovePathByShell(ChunkDir + uploadID)
//	if !delChkRes {
//		fmt.Printf("Failed to delete chunks as upload canceled, uploadID:%s\n", uploadID)
//	}
//
//	// 5. 响应客户端
//	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
//}
