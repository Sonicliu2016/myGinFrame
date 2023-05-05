package router

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	gj "github.com/segmentio/objconv/json"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"myGinFrame/controller"
	"myGinFrame/docs"
	"myGinFrame/glog"
	"myGinFrame/model"
	"myGinFrame/service"
	"myGinFrame/tool"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

type pathParam struct {
	keyName  string
	dataType reflect.Type
}

type Func struct {
	reflectVal     reflect.Value
	reflectValType reflect.Type
	pathParams     []*pathParam //path参数
	funcName       string
	controller     controller.Controller
	returnParams   []reflect.Type //返回值的参数类型
}

type response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"` //tag里面加上omitempy，可以在序列化的时候忽略0值或者空值
}

func InitRouter() *gin.Engine {
	// 创建gin路由,已经默认中间件（logger 和 recovery 中间件）
	//app := gin.Default()
	app := gin.New()
	app.Use(gin.Recovery())
	// 设置跨域
	app.Use(Cors())
	// 处理404找不到页面问题
	//app.NoRoute(notFindPage())
	// 设置swagger
	docs.SwaggerInfo.Title = "自定义Gin框架"
	docs.SwaggerInfo.Description = "自定义Gin框架，包含路由、mysql、mongodb、定时任务等封装"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	docs.SwaggerInfo.Host = tool.GetConfigStr("http_host") + ":" + tool.GetConfigStr("http_port")
	docs.SwaggerInfo.BasePath = "/v1"
	//url := ginSwagger.URL("http://" + tool.GetConfigStr("http_host") + ":" + tool.GetConfigStr("http_port") + "/swagger/doc.json") //The url pointing to API definition
	app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	//设置路由
	userService := service.NewUserService()
	userController := &controller.UserController{Service: userService}
	fileService := service.NewFileService()
	fileController := &controller.FileController{Service: fileService}
	rg := app.Group("/v1")
	{
		//	//v1.POST("/user", userController.NewUser)
		//	//v1.GET("/user/:userId", userController.GetUser)
		//	//v1.DELETE("/user/:userId", userController.DelUser)
		//	//v1.PUT("/user/:userId", userController.UpdateUser)
		setRouter(rg, userController)
		setRouter(rg, fileController)
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

//处理404，档找不到页面时，指定跳转到html/index.html页面
func notFindPage() gin.HandlerFunc {
	return func(context *gin.Context) {
		accept := context.Request.Header.Get("Accept")
		flag := strings.Contains(accept, "text/html")
		if flag {
			content, err := ioutil.ReadFile("html/index.html")
			if (err) != nil {
				context.Writer.WriteHeader(404)
				context.Writer.WriteString("Not Found")
				return
			}
			context.Writer.WriteHeader(200)
			context.Writer.Header().Add("Accept", "text/html")
			context.Writer.Write(content)
			context.Writer.Flush()
		}
	}
}

//TypeOf:动态的获取从函数接口中传进去的变量的类型,如果为空则返回值为nil（获取类型对象）
//->可以从该方法获取的对象中拿到字段的所属类型,字段名,以及该字段是否是匿名字段等信息
//->还可以获取到与该字段进行绑定的tag
//ValueOf:获取从函数入口传进去变量的值（获取值对象）
//->该方法获取到的对象是对应的数据信息区域(具体的值可以通过Filed(index)直接拿到对应位置的值)
//->FieldByName（字段名）
//->可以通过该函数拿到成员属性的详细信息(返回值与ValueOf一样)
//Elem()方法：该方法只接受指针类型与接口类型,如果不是这两种类型就会抛出异常（相当于对指针进行取元素,）
//原文链接：https://blog.csdn.net/apple_51931783/article/details/122478170
func setRouter(r *gin.RouterGroup, controller controller.Controller) {
	workPath, err := os.Getwd()
	if err != nil {
		glog.Glog.Error("Getwd err:", err)
		return
	}
	reflectVal := reflect.ValueOf(controller)
	reflectValType := reflect.TypeOf(controller)
	t := reflect.Indirect(reflectVal).Type()
	//获取当前文件夹(package的绝对路径)
	pkgRealpath := filepath.Join(workPath, "..", t.PkgPath())
	//glog.Glog.Info("workPath:", workPath, "-->t.PkgPath:", t.PkgPath(), "-->pkgRealpath:", pkgRealpath)
	structPath := path.Join(pkgRealpath, t.Name()) + ".go"
	//glog.Glog.Info("structPath:", structPath)
	//glog.Glog.Info("reflectValType.String():", reflectValType.String())
	//获取controller.UserController中的user字符串，用作url中的path
	//pathName := strings.ToLower(strings.ReplaceAll(strings.Split(reflectValType.String(), ".")[1], "Controller", ""))
	fileSet := token.NewFileSet()
	astPkgs, err := parser.ParseDir(fileSet, pkgRealpath, func(info os.FileInfo) bool {
		name := info.Name()
		return !info.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
	}, parser.ParseComments)
	if err != nil {
		glog.Glog.Error("ParseDir err:", err, "->pkgRealpath:", pkgRealpath)
		return
	}
	docFunMap := make(map[string]*ast.FuncDecl)
	for _, pkg := range astPkgs {
		for controllerName, fl := range pkg.Files {
			if strings.ToLower(controllerName) == strings.ToLower(structPath) {
				//glog.Glog.Info("controllerFileName:", controllerName)
				for _, d := range fl.Decls {
					switch d.(type) {
					//判断ast分类(获取注释信息，在注释中获取路由url以及action)
					case *ast.FuncDecl:
						demo := d.(*ast.FuncDecl)
						if demo.Name != nil && demo.Doc != nil {
							//glog.Glog.Info("funName:", demo.Name.Name)
							docFunMap[demo.Name.Name] = demo
						}
					}
				}
			}
		}
	}
	methodNum := reflectValType.NumMethod()
	for i := 0; i < methodNum; i++ {
		method := reflectValType.Method(i)
		//循环处理注册controller下面的方法
		if demo, ok := docFunMap[method.Name]; ok {
			for _, l := range demo.Doc.List {
				//获取注解中的路由path参数
				if strings.Contains(l.Text, "Router") {
					fun := &Func{
						reflectVal:     reflectVal,
						reflectValType: reflectValType,
						pathParams:     make([]*pathParam, 0),
						funcName:       method.Name,
						controller:     controller,
						returnParams:   make([]reflect.Type, 0),
					}
					routerParams := strings.Split(l.Text, " ")
					if len(routerParams) == 4 {
						//NumIn：方法参数个数
						//In：方法第 i 个参数的类型，i 的范围是 [0, NumIn() - 1]
						//NumOut：方法返回值个数
						//Out：方法第 i 个 返回值的类型，i 的范围为 [0, NumOut() - 1]
						//IsVariadic：判断方法是否存在可变参数
						funType := method.Func.Type()
						methodReturnNum := funType.NumOut()
						methodInNum := funType.NumIn()
						// "/image/{userId}/{faceSha1}" ->"/image/:userId/:faceSha1"
						routerPath := routerParams[2]
						pathParams := strings.Split(routerPath, "/")
						//router path中的key值(方法的参数)
						pathKeys := make([]string, 0)
						for _, pathKey := range pathParams {
							if strings.HasPrefix(pathKey, "{") && strings.HasSuffix(pathKey, "}") {
								pathKeys = append(pathKeys, pathKey[1:len(pathKey)-1])
							}
						}
						funParams := make([]reflect.Type, 0)
						for j := 0; j < methodInNum; j++ {
							in := funType.In(j)
							if in.Kind().String() != "ptr" && !strings.Contains(in.Name(), "controller") {
								funParams = append(funParams, in)
							}
						}
						// 要求path中的参数数量和方法中参数的数量保持一致
						if len(pathKeys) == len(funParams) {
							for i, key := range pathKeys {
								fun.pathParams = append(fun.pathParams, &pathParam{keyName: key, dataType: funParams[i]})
							}
							routerPath = strings.Replace(routerPath, "{", ":", -1)
							routerPath = strings.Replace(routerPath, "}", "", -1)
							action := strings.ToUpper(strings.Trim(strings.Trim(routerParams[3], "["), "]"))
							for j := 0; j < methodReturnNum; j++ {
								fun.returnParams = append(fun.returnParams, funType.Out(j))
							}
							//通过反射以及解析注解，拿到方法名以及传递的参数类型，实例化一个反射实例，call调用方法名并拿到返回值，然后进行http返回
							r.Handle(action, routerPath, fun.Handle())
						}
					}
				}
			}
		}
	}
}

func (fun *Func) Handle() func(ctx *gin.Context) {
	f := func(c *gin.Context) {
		defer c.Abort()
		params := make([]reflect.Value, 0)
		for _, pathParam := range fun.pathParams {
			pathValue := c.Param(pathParam.keyName)
			//将获取到的path值，转换为controller需要的类型参数
			switch pathParam.dataType.String() {
			case "int":
				if v, e := strconv.Atoi(pathValue); e == nil {
					params = append(params, reflect.ValueOf(v))
				}
			case "int64":
				if v, e := strconv.ParseInt(pathValue, 0, 64); e == nil {
					params = append(params, reflect.ValueOf(v))
				}
			case "string":
				params = append(params, reflect.ValueOf(pathValue))
			case "float32":
				if v, e := strconv.ParseFloat(pathValue, 32); e == nil {
					params = append(params, reflect.ValueOf(v))
				}
			case "float64":
				if v, e := strconv.ParseFloat(pathValue, 64); e == nil {
					params = append(params, reflect.ValueOf(v))
				}
			}
		}
		if len(params) == len(fun.pathParams) {
			//调用反射创建对象
			ptrValue := reflect.New(fun.reflectValType.Elem())
			//设置context
			ptrValue.Interface().(controller.Controller).SetContext(c)
			//设置service对象
			ptrValue.Elem().FieldByName("Service").Set(fun.reflectVal.Elem().FieldByName("Service"))
			//实例化一个反射对象，调用对象的相关方法，并传递参数
			returnValues := ptrValue.MethodByName(fun.funcName).Call(params)
			//c.Writer.Header().Set("Content-type", "application/json")
			//c.Writer.WriteHeader(200)
			//这里要求controller最多返回两个参数，需要返回的数据用map、interface{}或者结构体的方式，并且error放在最后一个位置
			if len(returnValues) > 0 && len(returnValues) == len(fun.returnParams) {
				for i, returnParam := range fun.returnParams {
					if returnParam.Name() == "error" {
						if i == 0 {
							switch returnValues[i].Interface().(type) {
							case model.NError:
								v := returnValues[i].Interface().(model.NError)
								//writeJson(c.Writer, response{Code: v.Code, Msg: v.Msg})
								sendJson(c, response{Code: v.Code, Msg: v.Msg})
								return
							case error:
								v := returnValues[i].Interface().(error)
								sendJson(c, fail(v.Error()))
								return
							default:
								sendJson(c, success(nil))
								return
							}
						} else {
							switch returnValues[i].Interface().(type) {
							case model.NError:
								v := returnValues[i].Interface().(model.NError)
								sendJson(c, response{Code: v.Code, Msg: v.Msg})
								return
							case error:
								v := returnValues[i].Interface().(error)
								sendJson(c, fail(v.Error()))
								return
							default:
								sendJson(c, success(returnValues[i-1].Interface()))
								return
							}
						}
					}
				}
				sendJson(c, success(returnValues[0].Interface()))
			} else {
				sendJson(c, success(nil))
			}
		} else {
			//说明参数转化时出错
			sendJson(c, fail("params input error!"))
		}
	}
	return f
}

func sendJson(ctx *gin.Context, obj response) {
	ctx.JSON(http.StatusOK, obj)
}

func writeJson(writer gin.ResponseWriter, obj response) {
	b, e := json.Marshal(obj)
	if e != nil {
		b, e = gj.Marshal(obj)
		if e != nil {
			b, _ = json.Marshal(response{Code: 1000, Msg: e.Error()})
		}
	}
	writer.Write(b)
}

func fail(msg string) response {
	return response{
		Code: 202,
		Msg:  msg,
	}
}

func success(data interface{}) response {
	return response{
		Code: 200,
		Msg:  "success",
		Data: data,
	}
}
