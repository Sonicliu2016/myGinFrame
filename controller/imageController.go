package controller

import (
	"github.com/gin-gonic/gin"
	"myGinFrame/service"
	"net/http"
)

type ImageController struct {
	BaseController
	Service service.UserServie
}

func (c *ImageController) GetImage(ctx *gin.Context) {
	userId, _ := c.Ctx.GetPostForm("userId")
	u := c.Service.GetUser(userId)
	c.Ctx.JSON(http.StatusOK, gin.H{
		"msg":  "ok",
		"data": u,
	})
}
