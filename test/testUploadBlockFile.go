package test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"myGinFrame/glog"
	"myGinFrame/model"
	"myGinFrame/net"
	"myGinFrame/tool"
	"net/http"
	"os"
	"strconv"
)

func Init() {
	glog.Glog.Info("测试分块上传!")
	filePath := "/home/liusong/go/go_project/r50.pth"
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
	// 2. 请求分块上传接口
	var chunksToUpload []int
	for idx := 1; idx <= result.Data.BlockCount; idx++ {
		chunksToUpload = append(chunksToUpload, idx)
	}
	// 上传所有分块
	uploadBlocks(result.Data.UploadId, filePath, result.Data.BlockSize, chunksToUpload)

}

// 实际上传分块逻辑
// chunkIdxs : 实际需要上传的分块
func uploadBlocks(uploadId, filePath string, blockSize int, blockIndexs []int) error {
	targetURL := "http://10.5.11.133:8080/v1/file/startUploadBlock" + "?uploadId=" + uploadId
	fh, err := os.Open(filePath)
	if err != nil {
		glog.Glog.Error("openfile err:", err)
		return err
	}
	defer fh.Close()

	reader := bufio.NewReader(fh)
	index := 0
	ch := make(chan int)
	buf := make([]byte, blockSize)
	for {
		n, err := reader.Read(buf) //每次读取blockSize大小的内容
		if n <= 0 {
			break
		}
		index++

		// 判断当前所在的块是否需要上传
		if contained, err := tool.Contain(blockIndexs, index); err != nil || !contained {
			continue
		}

		// 可以不使用bufCopied, 直接传将buf slice作为参数传递(值传递)进去计算sha1
		bufCopied := make([]byte, 5*1024*1024)
		copy(bufCopied, buf)

		go func(b []byte, blockIdx int) {
			resp, err := http.Post(
				targetURL+"&blockIndex="+strconv.Itoa(blockIdx),
				"multipart/form-data",
				bytes.NewReader(b))
			if err != nil {
				glog.Glog.Error("err:", err)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				glog.Glog.Error("ReadAll err:", err)
			}
			glog.Glog.Info("body:", string(body))
			resp.Body.Close()

			ch <- blockIdx
		}(bufCopied[:n], index)

		//遇到任何错误立即返回，并忽略 EOF 错误信息
		if err != nil {
			if err == io.EOF {
				break
			} else {
				glog.Glog.Error(err.Error())
			}
		}
	}

	for idx := 0; idx < len(blockIndexs); idx++ {
		select {
		case res := <-ch:
			glog.Glog.Info("完成传输块index: %d\n", res)
		}
	}

	glog.Glog.Info("全部完成以下分块传输: %+v\n", blockIndexs)
	return nil
}
