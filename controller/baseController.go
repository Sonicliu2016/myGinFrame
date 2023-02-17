package controller

import "github.com/gin-gonic/gin"

type BaseController struct {
	Ctx *gin.Context
}

type Controller interface {
	SetContext(ctx *gin.Context)
}

func (c *BaseController) SetContext(ctx *gin.Context) {
	c.Ctx = ctx
}
