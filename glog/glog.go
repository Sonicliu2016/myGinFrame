package glog

import (
	"encoding/json"
	"time"
)

var Glog *BeeLogger
var timeFormat = "2006-01-02 15:04:05"

func init() {
	timeStr := time.Now().Format("2006-01-02")
	config := map[string]interface{}{"filename": "logs/" + timeStr + ".log", "color": true, "level": 7, "maxdays": 30}
	configByte, _ := json.Marshal(config)
	Glog = NewLogger(1000)
	//l.SetLogger(logs.AdapterFile, `{"filename":"project.log","level":7,"maxlines":0,"maxsize":0,"daily":true,"maxdays":10,"color":true}`)
	//设置控制台输出
	Glog.SetLogger(AdapterConsole, `{"level":7,"color":true}`)
	//设置文件写入
	Glog.SetLogger(AdapterFile, string(configByte))
	//输出文件名和行号
	Glog.EnableFuncCallDepth(true)
	//获取打印文件绝对路径
	Glog.enableFullFilePath = true
	//获取打印文件相对路径
	Glog.enableRelativePath = true
	//SetLogFuncCallDepth参数(0,1,2,3)
	//0:具体获取行数的地方(log.go:270 -> runtime.Caller)，
	//1:调用打印的地方(log.go:530 -> bl.writeMsg(lm))，
	//2:真正打印输出的位置(glog.Glog.Info的位置)
	//3:上层调用的位置
	Glog.SetLogFuncCallDepth(2)
	//异步输出日志
	Glog.Async(1e3)
}
