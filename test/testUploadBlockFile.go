package test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"myGinFrame/glog"
	"myGinFrame/model"
	"myGinFrame/net"
	"myGinFrame/tool"
	"net/http"
	"os"
	"path"
	"strconv"
)

func Init() {
	glog.Glog.Info("测试分块上传!")
	filePath := path.Join(tool.GetConfigStr("staticPath"), "test.png")
	fileHash, err := tool.ComputeSha1ByShell(filePath)
	if err != nil {
		glog.Glog.Error("ComputeSha1ByShell err:", err)
		return
	}
	fileSize, err := tool.ComputeFileSizeByShell(filePath)
	if err != nil {
		glog.Glog.Error("ComputeFileSizeByShell err:", err)
		return
	}
	glog.Glog.Info("fileHash:", fileHash, "->fileSize:", fileSize)
	// 1. 初始化分块上传
	url := "http://10.5.11.133:8080/v1/file/initBlockUpload"
	buf, err := net.HttpPostFiles(url, nil, nil, map[string]string{"username": "ls", "fileHash": fileHash, "fileSize": strconv.Itoa(fileSize)})
	if err != nil {
		glog.Glog.Error("HttpPostFiles err:", err)
		return
	}
	glog.Glog.Info("buf:", string(buf))
	var result struct {
		Code int                       `json:"code"`
		Msg  string                    `json:"msg"`
		Data model.BlockFileUploadInfo `json:"data"`
	}
	if err = json.Unmarshal(buf, &result); err != nil {
		glog.Glog.Error("json.Unmarshal err:", err)
		return
	}
	if result.Code != 200 {
		return
	}
	// 2. 分块上传
	err = uploadBlocks(filePath, result.Data)
	if err != nil {
		return
	}
	// 3.合并分块
	url = "http://10.5.11.133:8080/v1/file/completeBlockUpload"
	buf, err = net.HttpPostFiles(url, nil, nil, map[string]string{"uploadId": result.Data.UploadId, "fileName": "1.png"})
	if err != nil {
		glog.Glog.Error("HttpPostFiles err:", err)
		return
	}
	glog.Glog.Info("完成分块上传:", string(buf))
}

// 实际上传分块逻辑
// chunkIdxs : 实际需要上传的分块
func uploadBlocks(filePath string, fileInfo model.BlockFileUploadInfo) error {
	targetURL := "http://10.5.11.133:8080/v1/file/startUploadBlock" + "?uploadId=" + fileInfo.UploadId
	fh, err := os.Open(filePath)
	if err != nil {
		glog.Glog.Error("openfile err:", err)
		return err
	}
	defer fh.Close()
	reader := bufio.NewReader(fh)

	index := 0
	completedIdxCh := make(chan int)
	errCh := make(chan error)
	buf := make([]byte, fileInfo.BlockSize)
	for {
		n, err := reader.Read(buf) //每次读取blockSize大小的内容
		if n <= 0 {
			break
		}
		index++

		// 判断当前所在的块是否需要上传
		if contained := tool.Contain(fileInfo.CompletedBlockIdxs, index); contained {
			continue
		}

		// 可以不使用bufCopied, 直接传将buf slice作为参数传递(值传递)进去计算sha1
		bufCopied := make([]byte, 5*1024*1024)
		copy(bufCopied, buf)

		go func(buf []byte, blockIdx int) {
			blockHash := tool.Sha1(buf)
			glog.Glog.Info("上传分块:", blockIdx, "->hash:", blockHash)
			resp, err := http.Post(
				targetURL+"&blockIndex="+strconv.Itoa(blockIdx)+"&blockHash="+blockHash,
				"multipart/form-data",
				bytes.NewReader(buf))
			if err != nil {
				glog.Glog.Error("err:", err)
				errCh <- err
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				glog.Glog.Error("ReadAll err:", err)
				errCh <- err
			}
			//glog.Glog.Info("body:", string(body))
			resp.Body.Close()
			var result struct {
				Code int    `json:"code"`
				Msg  string `json:"msg"`
			}
			if err = json.Unmarshal(body, &result); err != nil {
				glog.Glog.Error("json.Unmarshal err:", err)
				errCh <- err
			}
			if result.Code != 200 {
				errCh <- errors.New(result.Msg)
			}
			completedIdxCh <- blockIdx
		}(bufCopied[:n], index)

		//遇到任何错误立即返回，并忽略 EOF 错误信息
		if err != nil {
			if err == io.EOF {
				break
			} else {
				glog.Glog.Error(err.Error())
				errCh <- err
			}
		}
	}

	completedIdxs := fileInfo.CompletedBlockIdxs
	for {
		select {
		case idx := <-completedIdxCh:
			glog.Glog.Info("完成传输块index:", idx)
			completedIdxs = append(completedIdxs, idx)
		case err := <-errCh:
			glog.Glog.Error("err:", err)
			return err
		default:
			if len(completedIdxs) == fileInfo.BlockCount {
				glog.Glog.Info("全部完成以下分块传输:", completedIdxs)
				return nil
			}
		}
	}
	return nil
}
