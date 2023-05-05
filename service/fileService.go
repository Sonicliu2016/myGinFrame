package service

import (
	"errors"
	"fmt"
	"io"
	"math"
	"myGinFrame/glog"
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
	InitialMultipartUpload(username, fileHash string, fileSize int) error
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
	ChunkDir = "/data/chunks/"
	// MergeDir : 合并后的文件所在目录
	MergeDir = "/data/merge/"
	// ChunkKeyPrefix : 分块信息对应的redis键前缀
	ChunkKeyPrefix = "MP_"
	// HashUpIDKeyPrefix : 文件hash映射uploadid对应的redis键前缀
	HashUpIDKeyPrefix = "HASH_UPID_"
)

// 分块上传 : 初始化信息
type MultipartUploadInfo struct {
	FileHash string
	FileSize int
	UploadId string
	//每一个分块的大小
	ChunkSize  int
	ChunkCount int
	// 已经上传完成的分块索引列表
	ChunkExists []int
}

// InitialMultipartUpload : 初始化分块上传
func (s *fileService) InitialMultipartUpload(username, fileHash string, fileSize int) error {
	var uploadId string
	var err error
	// 1. 通过文件hash判断是否断点续传，并获取uploadID
	if redis.CheckKey(HashUpIDKeyPrefix + fileHash) {
		uploadId, err = redis.GetString(HashUpIDKeyPrefix + fileHash)
		if err != nil {
			return err
		}
	}

	// 2.1 首次上传则新建uploadID
	// 2.2 断点续传则根据uploadID获取已上传的文件分块列表
	chunksExist := []int{}
	if uploadId == "" {
		uploadId = username + fmt.Sprintf("%x", time.Now().UnixNano())
	} else {
		chunks, err := redis.Values(rConn.Do("HGETALL", ChunkKeyPrefix+uploadId))
		if err != nil {
			return err
		}
		for i := 0; i < len(chunks); i += 2 {
			k := string(chunks[i].([]byte))
			v := string(chunks[i+1].([]byte))
			if strings.HasPrefix(k, "chkidx_") && v == "1" {
				// chkidx_6 -> 6
				chunkIdx, _ := strconv.Atoi(k[7:len(k)])
				chunksExist = append(chunksExist, chunkIdx)
			}
		}
	}

	// 3. 生成分块上传的初始化信息
	upInfo := MultipartUploadInfo{
		FileHash:    fileHash,
		FileSize:    fileSize,
		UploadId:    uploadId,
		ChunkSize:   5 * 1024 * 1024,                                       // 5MB
		ChunkCount:  int(math.Ceil(float64(fileSize) / (5 * 1024 * 1024))), //向上取整
		ChunkExists: chunksExist,
	}

	// 4. 将初始化信息写入到redis缓存
	if len(upInfo.ChunkExists) <= 0 {
		hkey := ChunkKeyPrefix + upInfo.UploadId
		rConn.Do("HSET", hkey, "chunkcount", upInfo.ChunkCount)
		rConn.Do("HSET", hkey, "filehash", upInfo.FileHash)
		rConn.Do("HSET", hkey, "filesize", upInfo.FileSize)
		rConn.Do("EXPIRE", hkey, 60*60*12)
		rConn.Do("SET", HashUpIDKeyPrefix+fileHash, upInfo.UploadId, "EX", 60*60*12)
	}
	return nil
}

// MultipartUpload : 上传文件分块
func MultipartUpload(uploadId, chunkHash, chunkIndex string, body io.ReadCloser) error {
	// 1. 获得文件句柄，用于存储分块内容
	fpath := ChunkDir + uploadId + "/" + chunkIndex
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

	// 2. 校验分块hash (updated at 2020-05)
	cmpSha1, err := tool.ComputeSha1ByShell(fpath)
	if err != nil || cmpSha1 != chunkHash {
		glog.Glog.Error("Verify chunk sha1 failed, compare OK: %t, err:%+v\n", cmpSha1 == chunkHash, err)
		return err
	}

	// 3. 更新redis缓存状态
	rConn.Do("HSET", ChunkKeyPrefix+uploadId, "chkidx_"+chunkIndex, 1)
	return nil
}

// CompleteUploadHandler : 通知上传合并
func CompleteUploadHandler(uploadId, fileHash string) error {
	// 1. 通过uploadid查询redis并判断是否所有分块上传完成
	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+uploadId))
	if err != nil {
		return err
	}
	totalCount := 0
	chunkCount := 0
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount++
		}
	}
	if totalCount != chunkCount {
		glog.Glog.Info("totalCount:", totalCount, "->chunkCount:", chunkCount)
		return errors.New("totalCount != chunkCount")
	}

	// 4. 合并分块 (备注: 更新于2020/04/01; 此合并逻辑非必须实现，因后期转移到ceph/oss时也可以通过分块方式上传)
	if mergeSuc := tool.MergeChuncksByShell(ChunkDir+uploadId, MergeDir+fileHash, fileHash); !mergeSuc {
		return errors.New("complete upload failed")
	}

	// 5. 更新唯一文件表及用户文件表
	//fsize, _ := strconv.Atoi(fileSize)
	// 更新于2020-04: 增加fileaddr参数的写入
	//dblayer.OnFileUploadFinished(filehash, filename, int64(fsize), MergeDir+filehash)
	//dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(fsize))

	// 更新于2020-04: 删除已上传的分块文件及redis分块信息
	_, delHashErr := rConn.Do("DEL", HashUpIDKeyPrefix+fileHash)
	delUploadID, delUploadInfoErr := redis.Int64(rConn.Do("DEL", ChunkKeyPrefix+uploadId))
	if delUploadID != 1 || delUploadInfoErr != nil || delHashErr != nil {
		return errors.New("complete upload part failed")
	}

	delRes := tool.RemovePathByShell(ChunkDir + uploadId)
	if !delRes {
		fmt.Printf("Failed to delete chuncks as upload comoleted, uploadID: %s\n", uploadId)
	}
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
