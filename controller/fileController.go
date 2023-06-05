package controller

import (
	"myGinFrame/service"
	"myGinFrame/test"
	"strconv"
)

type FileController struct {
	BaseController
	Service service.FileServie
}

// @Tags 文件相关
// @Summary 文件上传
// @Description 文件上传
// @Accept  multipart/form-data
// @Produce json
// @Param   file  formData file true "文件"
// @Success 200 {string} Helloworld
// @Router /file/upload [post]
func (c *FileController) Upload() (string, error) {
	fh, err := c.Ctx.FormFile("file")
	if err != nil {
		return "", err
	}
	return c.Service.Upload(fh)
}

// @Tags 文件相关
// @Summary 测试分块上传
// @Description 测试分块上传
// @Accept  multipart/form-data
// @Produce json
// @Success 200 {string} Helloworld
// @Router /file/test [post]
func (c *FileController) Test() error {
	test.Init()
	return nil
}

// @Tags 文件相关
// @Summary 初始化分块上传
// @Description 初始化分块上传
// @Accept  multipart/form-data
// @Produce json
// @Param username formData string true  "用户姓名"
// @Param fileHash formData string false "文件hash值"
// @Param fileSize formData string false "文件大小"
// @Success 200 {string} Helloworld
// @Router /file/initBlockUpload [post]
func (c *FileController) InitBlockUpload() (interface{}, error) {
	username, _ := c.Ctx.GetPostForm("username")
	fileHash, _ := c.Ctx.GetPostForm("fileHash")
	fileSizeStr, _ := c.Ctx.GetPostForm("fileSize")
	fileSize, err := strconv.Atoi(fileSizeStr)
	if err != nil {
		return nil, err
	}
	info, err := c.Service.InitBlockUpload(username, fileHash, fileSize)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// @Tags 文件相关
// @Summary 分块上传
// @Description 分块上传
// @Accept  multipart/form-data
// @Produce json
// @Param username query string true  "用户姓名"
// @Param fileHash query string false "文件hash值"
// @Param fileSize query string false "文件大小"
// @Success 200 {string} Helloworld
// @Router /file/startUploadBlock [post]
func (c *FileController) StartUploadBlock() error {
	//uploadId, _ := c.Ctx.GetPostForm("uploadId")
	//blockHash, _ := c.Ctx.GetPostForm("blockHash")
	//blockIndex, _ := c.Ctx.GetPostForm("blockIndex")
	uploadId := c.Ctx.Query("uploadId")
	blockHash := c.Ctx.Query("blockHash")
	blockIndex := c.Ctx.Query("blockIndex")
	return c.Service.UploadBlock(uploadId, blockHash, blockIndex, c.Ctx.Request.Body)
}

// @Tags 文件相关
// @Summary 完成分块上传，开始合并
// @Description 完成分块上传，开始合并
// @Accept  multipart/form-data
// @Produce json
// @Param username formData string true  "用户姓名"
// @Param fileHash formData string false "文件hash值"
// @Param fileSize formData string false "文件大小"
// @Success 200 {string} Helloworld
// @Router /file/completeBlockUpload [post]
func (c *FileController) CompleteBlockUpload() error {
	uploadId, _ := c.Ctx.GetPostForm("uploadId")
	fileName, _ := c.Ctx.GetPostForm("fileName")
	return c.Service.CompleteUpload(uploadId, fileName)
}
