package net

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"myGinFrame/glog"
	"net/http"
	"os"
	"path"
	"time"
)

func HttpGet(url string, headers map[string]string, params map[string]string) ([]byte, error) {
	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		glog.Glog.Error("http.NewRequest err:", err)
		return nil, err
	}
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.Header.Add("Content-type", "application/json;charset=utf-8")
	req.URL.RawQuery = q.Encode()
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		glog.Glog.Error("client Do err:", err)
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// post上传文件
func HttpPostFile(url, filePath string, headers map[string]string) ([]byte, error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	//关键的一步操作
	fileWriter, err := bodyWriter.CreateFormFile("file", path.Base(filePath))
	if err != nil {
		glog.Glog.Error("CreateFormFile err:", err)
		return nil, err
	}
	//打开文件句柄操作
	fh, err := os.Open(filePath)
	if err != nil {
		glog.Glog.Error("open file err:", err)
		return nil, err
	}
	defer fh.Close()

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		glog.Glog.Error("Copy file err:", err)
		return nil, err
	}
	bodyWriter.Close()
	req, err := http.NewRequest(http.MethodPost, url, bodyBuf)
	if err != nil {
		glog.Glog.Error("http.NewRequest err:", err)
		return nil, err
	}
	req.Header.Add("Content-type", bodyWriter.FormDataContentType())
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	client := http.Client{Timeout: time.Minute * 5}
	resp, err := client.Do(req)
	if err != nil {
		glog.Glog.Error("client Do err:", err)
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// post发送raw格式的json
func HttpPostJson(url string, headers map[string]string, jsonData map[string]interface{}) ([]byte, error) {
	bytesData, err := json.Marshal(jsonData)
	if err != nil {
		glog.Glog.Error("json.Marshal err:", err)
		return nil, err
	}
	reader := bytes.NewReader(bytesData)
	req, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		glog.Glog.Error("http.NewRequest err:", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	client := http.Client{Timeout: time.Minute * 2}
	resp, err := client.Do(req)
	if err != nil {
		glog.Glog.Error("client Do err:", err)
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// post携带参数上传多文件
func HttpPostFiles(url string, files map[string]string, headers map[string]string, params map[string]string) ([]byte, error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	for k, v := range params {
		bodyWriter.WriteField(k, v)
	}
	for k, filePath := range files {
		fileWriter, err := bodyWriter.CreateFormFile(k, path.Base(filePath))
		if err != nil {
			glog.Glog.Error("CreateFormFile err:", err)
			return nil, err
		}
		//打开文件句柄操作
		fh, err := os.Open(filePath)
		if err != nil {
			glog.Glog.Error("open file err:", err)
			return nil, err
		}
		defer fh.Close()
		//iocopy
		_, err = io.Copy(fileWriter, fh)
		if err != nil {
			glog.Glog.Error("Copy file err:", err)
			return nil, err
		}
	}
	bodyWriter.Close()
	req, err := http.NewRequest(http.MethodPost, url, bodyBuf)
	if err != nil {
		glog.Glog.Error("http.NewRequest err:", err)
		return nil, err
	}
	req.Header.Set("Content-type", bodyWriter.FormDataContentType())
	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	client := http.Client{Timeout: time.Minute * 5}
	resp, err := client.Do(req)
	if err != nil {
		glog.Glog.Error("client Do err:", err)
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
