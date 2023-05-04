package redis

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"myGinFrame/glog"
	"strings"
	"sync"
)

type PSubscribeCallback func(pattern, channel, message string)

type PSubscriber struct {
	client redis.PubSubConn
	cbMap  map[string]PSubscribeCallback
	sync.RWMutex
}

func (c *PSubscriber) PConnect(conn redis.Conn) {
	if e := conn.Err(); e != nil {
		glog.Glog.Error("conn err:", e)
		return
	}

	c.client = redis.PubSubConn{conn}
	c.cbMap = make(map[string]PSubscribeCallback)

	go func() {
		for {
			//glog.Glog.Info("PSubscribe wait")
			switch res := c.client.Receive().(type) {
			case redis.Message:
				//可以是通配符，监听的是哪些key
				pattern := res.Pattern
				//具体返回哪个key
				channel := res.Channel
				key := string(res.Data)
				//调用回调函数
				if fun, ok := c.cbMap[channel]; ok {
					fun(pattern, channel, key)
				}
			case redis.Subscription:
				glog.Glog.Info("Subscription channel:", res.Channel, "-->Kind:", res.Kind, "-->Count:", res.Count)
			case error:
				glog.Glog.Error("PSubscribe error:", error.Error(res))
				continue
			}
		}
	}()
}

func (c *PSubscriber) Psubscribe(channel interface{}, cb PSubscribeCallback) {
	err := c.client.PSubscribe(channel)
	if err != nil {
		glog.Glog.Info("redis Subscribe error.")
	}
	c.Lock()
	c.cbMap[channel.(string)] = cb
	c.Unlock()
}

func PubCallback(patter, chann, key string) {
	//打开redis.conf文件，放开notify-keyspace-events "Ex"这行注释
	////PSUBSCRIBE __keyevent@index__:expired (index代表是哪个数据库,0-9)
	//glog.Glog.Info("patter:", patter, "->chann:", chann, "->key:", key)
	//__keyevent@1__:expired
	if chann == fmt.Sprintf("__keyevent@%d__:expired", "account_lock_countdown_") {
		if strings.HasPrefix(key, "account_lock_countdown_") {
			glog.Glog.Info("删除key:", key)
			strs := strings.SplitAfter(key, "account_lock_countdown_")
			if len(strs) == 2 {
				DelKey("account_lock_countdown_" + strs[1])
			}
		}
	}
}
