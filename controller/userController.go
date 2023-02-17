package controller

import (
	"github.com/gin-gonic/gin"
	"myGinFrame/service"
	"net/http"
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
func (c *UserController) NewUser(ctx *gin.Context) {
	username, _ := ctx.GetPostForm("username")
	tel, _ := ctx.GetPostForm("tel")
	gender, _ := ctx.GetPostForm("gender")
	err := c.Service.NewUser(username, tel, gender)
	if err == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"msg": "ok",
		})
	}
}

func (c *UserController) GetUser(ctx *gin.Context) {
	userId, _ := c.Ctx.GetPostForm("userId")
	u := c.Service.GetUser(userId)
	c.Ctx.JSON(http.StatusOK, gin.H{
		"msg":  "ok",
		"data": u,
	})
}

func (c *UserController) DelUser(ctx *gin.Context) {
	userId, _ := c.Ctx.GetPostForm("userId")
	err := c.Service.DeleteUser(userId)
	if err == nil {
		c.Ctx.JSON(http.StatusOK, gin.H{
			"msg": "ok",
		})
	}
}

func (c *UserController) UpdateUser(ctx *gin.Context) {
	userId, _ := c.Ctx.GetPostForm("userId")
	username, _ := c.Ctx.GetPostForm("username")
	tel, _ := c.Ctx.GetPostForm("tel")
	gender, _ := c.Ctx.GetPostForm("gender")
	err := c.Service.UpdateUser(userId, username, tel, gender)
	if err == nil {
		c.Ctx.JSON(http.StatusOK, gin.H{
			"msg": "ok",
		})
	}
}
