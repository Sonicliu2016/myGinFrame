package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"myGinFrame/controller"
	"myGinFrame/docs"
	"myGinFrame/service"
	"myGinFrame/tool"
	"net/http"
)

func InitRouter() *gin.Engine {
	// 创建gin路由,已经默认中间件（logger 和 recovery 中间件）
	//app := gin.Default()
	app := gin.New()
	app.Use(gin.Recovery())
	// 设置跨域
	app.Use(Cors())
	// 设置swagger
	docs.SwaggerInfo.Title = "自定义Gin框架"
	docs.SwaggerInfo.Description = "自定义Gin框架，包含路由、mysql、mongodb、定时任务等封装"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = tool.GetConfigStr("http_host") + ":" + tool.GetConfigStr("http_port")
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	docs.SwaggerInfo.BasePath = "/v1"
	//url := ginSwagger.URL("http://" + tool.GetConfigStr("http_host") + ":" + tool.GetConfigStr("http_port") + "/swagger/doc.json") //The url pointing to API definition
	app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	//设置路由
	userService := service.NewUserService()
	userController := controller.UserController{Service: userService}
	v1 := app.Group("/v1")
	{
		v1.POST("/user", userController.NewUser)
		v1.GET("/user/:userId", userController.GetUser)
		v1.DELETE("/user/:userId", userController.DelUser)
		v1.PUT("/user/:userId", userController.UpdateUser)
	}
	return app
}

// 设置跨域
func Cors() gin.HandlerFunc {
	return func(context *gin.Context) {
		method := context.Request.Method
		context.Header("Access-Control-Allow-Origin", "*")
		context.Header("Access-Control-Allow-Headers", "Content-Type, AccessToken, X-CSRF-Token, Authorization, Token, x-token")
		context.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH, PUT")
		context.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		context.Header("Access-Control-Allow-Credentials", "true")
		if method == "OPTIONS" {
			context.AbortWithStatus(http.StatusNoContent)
		}
	}
}
