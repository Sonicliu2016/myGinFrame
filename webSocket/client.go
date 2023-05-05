package webSocket

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"myGinFrame/glog"
	"net/url"
	"time"
)

type websocketClientManager struct {
	conn                 *websocket.Conn
	addr                 *string
	path                 string
	rawQuery             string
	ReceiveMsgFromServer func(data []byte)
	sendMsgChan          chan []byte
	isAlive              bool
	timeout              int
}

// 构造函数
func NewWsClientManager(addrIp, addrPort, path, rawQuery string, timeout int) *websocketClientManager {
	addrString := addrIp + ":" + addrPort
	var sendChan = make(chan []byte, 10)
	var conn *websocket.Conn
	return &websocketClientManager{
		addr:        &addrString,
		path:        path,
		rawQuery:    rawQuery,
		conn:        conn,
		sendMsgChan: sendChan,
		isAlive:     false,
		timeout:     timeout,
	}
}

// 链接服务端
func (wsc *websocketClientManager) dail() {
	var err error
	u := url.URL{Scheme: "ws", Host: *wsc.addr, Path: wsc.path, ForceQuery: true, RawQuery: wsc.rawQuery}
	glog.Glog.Info("connecting to %s", u.String())
	wsc.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		glog.Glog.Error("dial err:", err)
		return
	}
	wsc.isAlive = true
	glog.Glog.Info("connecting to %s 链接成功", u.String())
}

// 发送消息
func (wsc *websocketClientManager) sendMsgThread() {
	go func() {
		for {
			msg := <-wsc.sendMsgChan
			//glog.Glog.Info("send msg:", string(msg))
			err := wsc.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				glog.Glog.Error("write msg err:", err)
				continue
			}
		}
	}()
}

// 读取消息
func (wsc *websocketClientManager) readMsgThread() {
	go func() {
		for {
			if wsc.conn != nil {
				_, message, err := wsc.conn.ReadMessage()
				if err != nil {
					glog.Glog.Error("read msg err:", err)
					wsc.isAlive = false
					// 出现错误，退出读取，尝试重连
					break
				}
				//glog.Glog.Info("recv msg:", string(message))
				if wsc.ReceiveMsgFromServer != nil {
					wsc.ReceiveMsgFromServer(message)
				}
			}
		}
	}()
}

// 开启服务并重连
func (wsc *websocketClientManager) Start() {
	for {
		if wsc.isAlive == false {
			wsc.dail()
			wsc.sendMsgThread()
			wsc.readMsgThread()
		}
		ping := map[string]interface{}{
			"header": map[string]int{
				"type": 1,
			},
			"body": "ping",
		}
		bytesData, _ := json.Marshal(ping)
		wsc.SendMsg(bytesData)
		time.Sleep(time.Second * time.Duration(wsc.timeout))
	}
}

func (wsc *websocketClientManager) SendMsg(data []byte) {
	//glog.Glog.Info("send msg chan:", string(data))
	wsc.sendMsgChan <- data
}
