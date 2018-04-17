package utils

import (
	"runtime"

	log "github.com/Sirupsen/logrus"
	//"github.com/davecgh/go-spew/spew"
)

// 产生panic时的调用栈打印
func PrintPanicStack(extras ...interface{}) {
	if x := recover(); x != nil {
		log.WithFields(log.Fields{
			"err": x,
		}).Error("print panic stack")

		// 协议错误不需要记录调用栈信息
		switch t := x.(type) {
		case string:
			if t == "error occured in protocol module" {
				return
			}
		}

		i := 0
		funcName, file, line, ok := runtime.Caller(i)
		for ok {
			log.Errorf("panic frame %v:[func:%v,file_name:%v,file_line:%v]\n", i, runtime.FuncForPC(funcName).Name(), file, line)
			i++
			funcName, file, line, ok = runtime.Caller(i)
		}

		//for k := range extras {
		//	log.Errorf("panic EXRAS#%v DATA:%v\n", k, spew.Sdump(extras[k]))
		//}
	}
}
