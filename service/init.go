package service

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"myGinFrame/glog"
	"net/url"
	"os"
	"os/signal"
)

func init() {
	glog.Glog.Info("service init")
	//websocketClient()
}

var addr = flag.String("addr", "127.0.0.1:8085", "http service address")

func websocketClient() {
	flag.Parse()
	log.SetFlags(0)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws", ForceQuery: true, RawQuery: "name=mvlite"}
	glog.Glog.Info("connecting url:", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		glog.Glog.Error("dial err:", err)
	}
	defer c.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				glog.Glog.Error("read msg err:", err)
				return
			}
			glog.Glog.Info("recv msg:", string(message))
		}
	}()

	data := map[string]interface{}{
		"header": map[string]int{
			"type": 1,
		},
		"body": "ping",
	}
	bytesData, _ := json.Marshal(data)
	err = c.WriteMessage(websocket.TextMessage, bytesData)
	glog.Glog.Info("send heart msg:", string(bytesData))

	data = map[string]interface{}{
		"header": map[string]int{
			"type": 3,
		},
		"filePath":    "/home/liusong/go/mm_models.zip",
		"storagePath": "/edgeai.com/mvlite-models",
	}
	bytesData, _ = json.Marshal(data)
	err = c.WriteMessage(websocket.TextMessage, bytesData)
	glog.Glog.Info("send file msg:", string(bytesData))

	//ticker := time.NewTicker(time.Second * 5)
	//defer ticker.Stop()
	//for {
	//	select {
	//	case <-done:
	//		return
	//	case t := <-ticker.C:
	//		data := map[string]interface{}{
	//			"header": map[string]int{
	//				"type": 1,
	//			},
	//			"body": "ping",
	//		}
	//		bytesData, err := json.Marshal(data)
	//		if err != nil {
	//			glog.Glog.Error("json.Marshal err:", err)
	//			return
	//		}
	//		err = c.WriteMessage(websocket.TextMessage, bytesData)
	//		glog.Glog.Info("send msg:", t.String())
	//		if err != nil {
	//			glog.Glog.Error("write msg err:", err)
	//			return
	//		}
	//	case <-interrupt:
	//		glog.Glog.Info("interrupt")
	//		// Cleanly close the connection by sending a close message and then
	//		// waiting (with timeout) for the server to close the connection.
	//		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	//		if err != nil {
	//			glog.Glog.Error("WriteMessage err:", err)
	//			return
	//		}
	//		select {
	//		case <-done:
	//		case <-time.After(time.Second):
	//		}
	//		return
	//	}
	//}

}
