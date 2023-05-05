package controller

import (
	"myGinFrame/glog"
	"myGinFrame/model"
	"myGinFrame/service"
)

type UserController struct {
	BaseController
	Service service.UserServie
}

// @Tags 用户相关
// @Summary 添加用户
// @Description 添加用户
// @Accept  multipart/form-data
// @Produce json
// @Param username formData string true  "用户姓名"
// @Param tel      formData string false "用户Id(不可修改，不可重复。支持英文、数字以及特殊字符),如果用户没有指定，则系统自动生成"
// @Param gender   formData string false "用户性别（0代表未填写；1代表男性；2代表女性)"
// @Success 200 {string} Helloworld
// @Router /user [post]
func (c *UserController) NewUser() error {
	username, _ := c.Ctx.GetPostForm("username")
	tel, _ := c.Ctx.GetPostForm("tel")
	gender, _ := c.Ctx.GetPostForm("gender")
	err := c.Service.NewUser(username, tel, gender)
	return err
}

// @Tags 用户相关
// @Summary 添加用户图像
// @Description 添加用户图像
// @Accept  multipart/form-data
// @Produce json
// @Param userId path     int    true "用户id"
// @Param image  formData string true "头像，base64"
// @Success 200 {string} Helloworld
// @Router /user/{userId}/headImage [post]
func (c *UserController) UserImage(userId int) error {
	image, _ := c.Ctx.GetPostForm("image")
	glog.Glog.Info("HeadImage userId:", userId, "->image:", image)
	return nil
}

// @Tags 用户相关
// @Summary 查询用户
// @Description 查询用户
// @Accept  multipart/form-data
// @Produce json
// @Param   userId path string true "用户id"
// @Success 200 {string} Helloworld
// @Router /user/{userId} [get]
func (c *UserController) GetUser(userId string) interface{} {
	u := c.Service.GetUser(userId)
	//r,err := tool.RunShell("go","env")
	//glog.Glog.Info("r:", r,"->err:",err)
	return u
}

// @Tags 用户相关
// @Summary 删除用户
// @Description 删除用户
// @Accept  multipart/form-data
// @Produce json
// @Param   userId path string true "用户id"
// @Success 200 {string} Helloworld
// @Router /user/{userId} [delete]
func (c *UserController) DelUser(userId string) {
	glog.Glog.Info("DelUser userId:", userId)
	c.Service.DeleteUser(userId)
	return
}

// @Tags 用户相关
// @Summary 更新用户
// @Description 更新用户
// @Accept  multipart/form-data
// @Produce json
// @Param   userId   path     int    true "用户id"
// @Param   faceSha1 path     string true "hash值"
// @Param   username formData string true "username"
// @Success 200 {string} Helloworld
// @Router /user/{userId}/{faceSha1} [put]
func (c *UserController) UpdateUser(userId int, faceSha1 string) (int, string, error) {
	username, _ := c.Ctx.GetPostForm("username")
	glog.Glog.Info("UpdateUser userId:", userId, "->faceSha1:", faceSha1, "->username:", username)
	c.Service.UpdateUser("", username, "", "")
	return 0, "shangsan", model.NError{Code: 403, Msg: "气死你"}
}
