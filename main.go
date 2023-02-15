package main

import (
	"myGinFrame/glog"
	"net/http"
)

func main() {
	glog.Glog.Info("我是打印信息", "11111111111111111111111111111")
	glog.Glog.Error("我是打印信息", "2222222222222222222222222222")
	glog.Glog.Alert("我是打印信息", "3333333333333333333333333333")
	glog.Glog.Debug("我是打印信息", "4444444444444444444444444444")
	http.ListenAndServe(":8080", nil)
}
