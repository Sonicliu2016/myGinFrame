package glog

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

type LogMsg struct {
	Level               int
	Msg                 string
	When                time.Time
	FilePath            string
	LineNumber          int
	Args                []interface{}
	Prefix              string
	enableFullFilePath  bool
	enableRelativePath  bool
	enableFuncCallDepth bool
}

// 对消息进行一个格式化输出
func (lm *LogMsg) OldStyleFormat() string {
	msg := lm.Msg

	if len(lm.Args) > 0 {
		//msg = fmt.Sprintf(lm.Msg, lm.Args...)
		for _, arg := range lm.Args {
			msg = fmt.Sprintf("%v%v", msg, arg)
		}
	}

	msg = lm.Prefix + " " + msg
	//如果设置了获取文件名和行号
	if lm.enableFuncCallDepth {
		filePath := lm.FilePath
		//如果设置了获取全路径
		if lm.enableFullFilePath {
			if lm.enableRelativePath {
				//获取项目路径
				rootPath, err := os.Getwd()
				if err == nil {
					filePath = strings.Replace(filePath, rootPath+"/", "", -1)
				}
			}
		} else {
			_, filePath = path.Split(filePath)
		}
		//msg = fmt.Sprintf("[%s:%d] %s", filePath, lm.LineNumber, msg)
		//按照 "%s:%d:" 的格式才会输出带链接
		msg = fmt.Sprintf("%s:%d:%s", filePath, lm.LineNumber, msg)
	}

	msg = levelPrefix[lm.Level] + " " + msg
	return msg
}
