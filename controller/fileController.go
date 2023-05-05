package controller

import (
	"myGinFrame/service"
	"strconv"
)

type FileController struct {
	BaseController
	Service service.FileServie
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
