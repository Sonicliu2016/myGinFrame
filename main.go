package main

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"myGinFrame/glog"
	"myGinFrame/router"
	"myGinFrame/tool"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard
	gin.ForceConsoleColor()
	app := router.InitRouter()
	app.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	if err := app.Run(":" + tool.GetConfigStr("http_port")); err != nil {
		glog.Glog.Error("app.Run err:", err)
	}
}
