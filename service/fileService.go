package service

import (
	"errors"
	"fmt"
	"gopkg.in/amz.v1/s3"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"myGinFrame/ceph"
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
	Upload(fh *multipart.FileHeader) (string, error)
	//InitBlockUpload : 初始化分块上传
	InitBlockUpload(username, fileHash string, fileSize int) (interface{}, error)
	//UploadBlock : 上传文件分块
	UploadBlock(uploadId, blockHash, blockIndex string, body io.ReadCloser) error
	//CompleteUpload : 通知上传合并
	CompleteUpload(uploadId, fileName string) error
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
	// HashUpIDKeyPrefix : 文件hash映射uploadid对应的redis键前缀
	UploadIdKeyPrefix = "upload_id_"
	// fileKeyPrefix : 分块信息对应的redis键前缀
	FileKeyPrefix    = "file_"
	BlockIndexPrefix = "blockIndex_"
)

func (s *fileService) Upload(fh *multipart.FileHeader) (string, error) {
	file, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()
	bucket := ceph.GetCephBucket("file")
	// 创建一个新的bucket
	if err := bucket.PutBucket(s3.PublicRead); err != nil {
		glog.Glog.Error("bucket.PutBucket err:", err)
	}
	// 查询这个bucket下面指定条件的object keys
	res, err := bucket.List("", "", "", 100)
	if err != nil {
		glog.Glog.Error("bucket list err:", err)
	} else {
		glog.Glog.Info("object keys: %+v", res)
	}
	// 上传文件
	cephSavePath := "/ceph/test/"
	fileBody, err := ioutil.ReadAll(file)
	if err != nil {
		glog.Glog.Error("Read file failed, %+v\n", err)
		return "", err
	}
	if err = bucket.Put(cephSavePath, fileBody, "octet-stream", s3.PublicRead); err != nil {
		glog.Glog.Error("upload file err: %+v\n", err)
		return "", err
	}

	// 下载文件B
	objB, err := bucket.Get(cephSavePath)
	if err != nil {
		glog.Glog.Error("Get object B err: %s\n", err.Error())
		return "", err
	}
	tmpFile, err := os.Create(tool.GetConfigStr("staticPath") + "/" + fh.Filename + ".copy")
	if err != nil {
		glog.Glog.Error("Write object B to file err: %s\n", err.Error())
		return "", err
	}
	tmpFile.Write(objB)

	// 查询这个bucket下面指定条件的object keys
	res, err = bucket.List("", "", "", 100)
	if err != nil {
		glog.Glog.Error("bucket list err:", err.Error())
	} else {
		glog.Glog.Info("object keys:", res)
	}
	return cephSavePath, nil
}

// InitBlockUpload : 初始化分块上传
func (s *fileService) InitBlockUpload(username, fileHash string, fileSize int) (interface{}, error) {
	var uploadId string
	var err error
	// 1. 通过文件hash判断是否断点续传，并获取uploadId
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
			//glog.Glog.Info("k:", k, "->v:", v)
			if strings.HasPrefix(k, BlockIndexPrefix) && v == "ok" {
				// blockIndex_6 -> 6
				blockIdx, _ := strconv.Atoi(k[len(BlockIndexPrefix):])
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
func (s *fileService) CompleteUpload(uploadId, fileName string) error {
	// 1. 通过uploadid查询redis并判断是否所有分块上传完成
	blocks, err := redis.GetHashAll(FileKeyPrefix + uploadId)
	if err != nil {
		return err
	}
	totalCount := 0
	chunkCount := 0
	var fileHash string
	for i := 0; i < len(blocks); i += 2 {
		k := string(blocks[i].([]byte))
		v := string(blocks[i+1].([]byte))
		if k == "fileHash" {
			fileHash = v
		}
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

	// 2. 合并分块
	blockFolder := tool.GetConfigStr("staticPath") + BlockDir + uploadId
	mergeSavePath := tool.GetConfigStr("staticPath") + MergeDir + fileName
	os.MkdirAll(path.Dir(mergeSavePath), os.ModePerm)
	if mergeSuc := tool.MergeChuncksByShell(blockFolder, mergeSavePath, fileHash); !mergeSuc {
		return errors.New("complete upload failed")
	}

	// 删除已上传的分块文件及redis分块信息
	redis.DelKey(UploadIdKeyPrefix + fileHash)
	redis.DelKey(FileKeyPrefix + uploadId)
	os.RemoveAll(blockFolder)
	return nil
}
