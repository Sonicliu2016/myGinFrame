package tool

import (
	"fmt"
	"github.com/robfig/config"
)

var TimeFormat = "2006-01-02 15:04:05"

var c *config.Config
var err error

func init() {
	c, err = config.ReadDefault("./conf/conf.ini")
	if err != nil {
		fmt.Println("server fail:", err.Error())
	}
}

func GetConfigStr(key string) string {
	str, err := c.String("DEFAULT", key)
	if err == nil {
		return str
	} else {
		return ""
	}
}

func GetConfigInt(key string) int {
	i, err := c.Int("DEFAULT", key)
	if err == nil {
		return i
	} else {
		return 1
	}
}

func GetConfigFloat(key string) float64 {
	i, err := c.Float("DEFAULT", key)
	if err == nil {
		return i
	} else {
		return 1
	}
}
