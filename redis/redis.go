package redis

import (
	"encoding/json"
	"myGinFrame/glog"
	"myGinFrame/tool"
	"strconv"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	pool      *redis.Pool = nil
	MaxIdle   int         = 10
	MaxActive int         = 128
	ExpireSec int64       = 0
	lock      sync.RWMutex
)

func init() {
	InitRedis()
}

func InitRedis() {
	pool = newRedisPool()
	var psub PSubscriber
	psub.PConnect(pool.Get())
	//打开redis.conf文件，放开notify-keyspace-events "Ex"这行注释
	//PSUBSCRIBE __keyevent@index__:expired (index代表是哪个数据库,0-9)
	//__keyevent@1__:expired
	//channel := fmt.Sprintf("__keyevent@%d__:expired", "account_lock_countdown_")
	//psub.Psubscribe(channel, PubCallback)
}

func newRedisPool() *redis.Pool {
	glog.Glog.Info("初始化redis")
	redisAddress := tool.GetConfigStr("redisUrl")
	redisPwd := tool.GetConfigStr("redisPwd")
	MaxIdle = tool.GetConfigInt("redisMaxIdle")
	MaxActive = tool.GetConfigInt("redisMaxActive")
	if len(redisAddress) == 0 {
		redisAddress = "127.0.0.1:6379"
	}
	if len(redisPwd) == 0 {
		redisPwd = "ailab"
	}
	pool := &redis.Pool{
		//最大的空闲连接数，表示即使没有redis连接时依然可以保持N个空闲的连接，而不被清除，随时处于待命状态
		MaxIdle: MaxIdle,
		//最大的激活连接数，表示同时最多有N个连接
		MaxActive: MaxActive,
		//最大的空闲连接等待时间，超过此时间后，空闲连接将被关闭
		IdleTimeout: 300 * time.Second,
		//建立连接
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp",
				redisAddress,
				//redis.DialReadTimeout(5*time.Second),
				redis.DialWriteTimeout(5*time.Second),
				redis.DialConnectTimeout(5*time.Second),
			)
			if err != nil {
				glog.Glog.Error("连接redis服务错误:", err)
				return nil, err
			}
			if redisPwd != "" {
				if _, authErr := c.Do("AUTH", redisPwd); authErr != nil {
					c.Close()
					glog.Glog.Error("验证redis密码错误:", authErr)
					return nil, authErr
				}

			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			if err != nil {
				glog.Glog.Error("ping redis error:", err)
				return err
			}
			return nil
		},
	}

	return pool
}

func rdsdo(cmd string, key string, args ...interface{}) (interface{}, error) {
	dbNum := 10
	//switch {
	//case strings.HasPrefix(key, model.EMAILCODE_COUNTDOWN_RESULT_KEY):
	//	dbNum = model.EMAILCODE_COUNTDOWN_DB
	//case strings.HasPrefix(key, model.ACCOUNT_LOCK_COUNTDOWN_RESULT_KEY):
	//	dbNum = model.ACCOUNT_LOCK_COUNTDOWN_DB
	//case strings.HasPrefix(key, model.TOKEN_EXPIRED_COUNTDOWN__RESULT_KEY):
	//	dbNum = model.TOKEN_EXPIRED_COUNTDOWN_DB
	//case strings.HasPrefix(key, model.LOGINPWD_ERROR_COUNT_RESULT_KEY):
	//	dbNum = model.LOGINPWD_ERROR_COUNT_DB
	//}
	conn := pool.Get()
	lock.RLock()
	defer lock.RUnlock()
	if err := conn.Err(); err != nil {
		return nil, err
	}
	defer conn.Close()
	conn.Do("SELECT", dbNum)
	parmas := make([]interface{}, 0)
	parmas = append(parmas, key)

	if len(args) > 0 {
		for _, v := range args {
			parmas = append(parmas, v)
		}
	}
	return conn.Do(cmd, parmas...)
}

func SetString(key string, value string) error {
	_, err := rdsdo("SET", key, value)
	if err != nil {
		glog.Glog.Error("set error", err.Error())
		return err
	}
	return nil
}

/*
 * 给定key设置过期时间，以秒计
 */
func SetKeyExpire(key string, ex int) error {
	_, err := rdsdo("EXPIRE", key, ex)
	if err != nil {
		glog.Glog.Error("set expire error:", err.Error())
		return err
	}
	return nil
}

/*
 * 获取key过期时间，以秒计
 */
func GetKeyExpire(key string) int {
	result, err := rdsdo("TTL", key)
	if err != nil {
		glog.Glog.Error("get expire error:", err.Error())
		return -3
	}
	expire, err := redis.Int(result, err)
	if err != nil {
		glog.Glog.Error("redis.Int:", err.Error())
		return -3
	}
	return expire
}

/*
 * 给定key设置值并且设置过期时间，以秒计
 */
func SetKeyAndExpire(key, value string, ex int) error {
	_, err := rdsdo("SET", key, value, "EX", ex)
	if err != nil {
		glog.Glog.Error("set value and expire error:", err.Error())
		return err
	}
	return nil
}

/*
 * 检查key是否存在
 */
func CheckKey(key string) bool {
	b, err := redis.Bool(rdsdo("EXISTS", key))
	if err != nil {
		glog.Glog.Error("CheckKey err:", err)
		return false
	}
	return b
}

/*
 * 获取key对应的value值
 */
func GetString(key string) (string, error) {
	result, err := rdsdo("GET", key)
	if nil == err {
		str, _ := redis.String(result, err)
		return str, nil
	} else {
		glog.Glog.Debug("redis get error:" + err.Error())
		return "", err
	}
}

func GetInt64(key string) int64 {
	//glog.Glog.Info("Get int key:",key)
	numStr, err := GetString(key)
	if err == nil {
		var num int64
		if len(numStr) > 0 {
			num, err = strconv.ParseInt(numStr, 10, 64)
			if err == nil {
				return num
			}
		}
	}
	return 0
}

/*
 * 删除key
 */
func DelKey(key string) error {
	_, err := rdsdo("DEL", key)
	if err != nil {
		glog.Glog.Error("DelKey err:", err)
		return err
	}
	return nil
}

/*
 * 将key中储存的数字值增一
 */
func VIncrease(key string) (int64, error) {
	i, err := rdsdo("INCR", key)
	if err != nil {
		glog.Glog.Error("VIncrease error", err.Error())
		return 0, err
	}
	return i.(int64), nil
}

func VIncreaseAndExpire(key string, second int) error {
	_, err := rdsdo("INCR", key)
	if err == nil {
		_, err = rdsdo("EXPIRE", key, second)
		if err != nil {
			glog.Glog.Info("EXPIRE err:", err)
		}
		return err
	} else {
		glog.Glog.Info("INCR err:", err)
	}
	return err
}

func VIncreaseN(key string, n int) error {
	_, err := rdsdo("INCRBY", key, n)
	if err != nil {
		glog.Glog.Error("VIncreaseN error", err.Error())
		return err
	}
	return nil
}

/**
 * 将key中储存的数字值减一。
 */
func VDecrease(key string) error {
	_, err := rdsdo("DECR", key)
	if err != nil {
		glog.Glog.Error("VDecrease error", err.Error())
		return err
	}
	return nil
}

/**
 * 将key所储存的值减去减量 n
 */
func VDecreaseN(key string, n int) error {
	_, err := rdsdo("DECRBY", key, n)
	if err != nil {
		glog.Glog.Error("VDecreaseN error", err.Error())
		return err
	}
	return nil
}

func WriteStruct(key string, obj interface{}) error {
	data, err := json.Marshal(obj)
	if nil == err {
		return SetString(key, string(data))
	} else {
		return err
	}
}

/*
 * 给定key设置值并且设置过期时间，以秒计
 */
func SetStructAndExpire(key string, obj interface{}, second int) error {
	data, err := json.Marshal(obj)
	if nil == err {
		return SetKeyAndExpire(key, string(data), second)
	} else {
		return err
	}
}

func ReadStruct(key string, obj interface{}) error {
	if data, err := GetString(key); nil == err {
		return json.Unmarshal([]byte(data), obj)
	} else {
		return err
	}
}

func WriteList(key string, list interface{}, total int) error {
	realKeyList := key + "_list"
	realKeyCount := key + "_count"
	data, err := json.Marshal(list)
	if nil == err {
		SetString(realKeyCount, strconv.Itoa(total))
		return SetString(realKeyList, string(data))
	} else {
		return err
	}
}

func ReadList(key string, list interface{}) (int, error) {
	realKeyList := key + "_list"
	realKeyCount := key + "_count"
	if data, err := GetString(realKeyList); nil == err {
		totalStr, _ := GetString(realKeyCount)
		total := 0
		if len(totalStr) > 0 {
			total, _ = strconv.Atoi(totalStr)
		}
		return total, json.Unmarshal([]byte(data), list)
	} else {
		return 0, err
	}
}

/**
redis SADD 将一个或多个 member 元素加入到集合 key 当中，已经存在于集合的 member 元素将被忽略。
*/
func RdbSAdd(key, v string) error {
	_, err := rdsdo("SADD", key, v)
	if err != nil {
		glog.Glog.Error("SADD error", err.Error())
		return err
	}
	return nil
}

/**
redis SMEMBERS 返回集合 key 中的所有成员。
return map
*/
func RdbSMembers(key string) (interface{}, error) {
	data, err := redis.Strings(rdsdo("SMEMBERS", key))
	if err != nil {
		glog.Glog.Error("json nil", err)
		return nil, err
	}
	return data, nil
}

/**
redis SISMEMBER 判断 member 元素是否集合 key 的成员。
return bool
*/
func RdbSISMembers(key, v string) bool {
	b, err := redis.Bool(rdsdo("SISMEMBER", key, v))
	if err != nil {
		glog.Glog.Error("SISMEMBER error", err.Error())
		return false
	}
	return b
}

// 存储一个对象
// HMSET people name "zhangsan" age 20 address "shenzhen"
func SetHash(key, k string, v interface{}) error {
	_, err := rdsdo("HSET", key, k, v)
	if err != nil {
		glog.Glog.Error("HSET error", err.Error())
		return err
	}
	return nil
}

// 获取一个对象
// HGETALL people
// "name" "zhangsan" "age" "20" "address" "shenzhen"
func GetHashAll(key string) ([]interface{}, error) {
	result, err := rdsdo("HGETALL", key)
	if nil != err {
		glog.Glog.Debug("redis HGETALL error:" + err.Error())
		return nil, err
	}
	return redis.Values(result, err)
}
