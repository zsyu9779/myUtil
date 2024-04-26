package go_behind

import (
	"codeup.aliyun.com/aha/social_aha_gotool/alarm_notify"
	"codeup.aliyun.com/aha/social_aha_gotool/gopool"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type GoBehindConfig struct {
	NamespaceMode string
	Logger        *logrus.Logger
	Pool          gopool.Pool
}

var goBehindCfg *GoBehindConfig

// InitGoBehind 初始化go behind 各业务端灵活选择是否init 支持传入namespace 自定义的goRoutine pool和logger
func InitGoBehind(cfg *GoBehindConfig) {
	goBehindCfg = cfg
}

// NOTICE: 建议使用InitGoBehind 如不使用InitGoBehind 则使用默认配置如下
func getConfig() *GoBehindConfig {
	if goBehindCfg == nil {
		goBehindCfg = &GoBehindConfig{
			NamespaceMode: os.Getenv("NAMESPACE_MODE"),
			Logger:        logrus.New(),
			Pool:          gopool.NewPool("default", 10000, gopool.NewConfig()),
		}
	}
	return goBehindCfg
}

func caller() (string, string) {
	pc, file, line, _ := runtime.Caller(3)
	function := runtime.FuncForPC(pc).Name()                    // 获取函数名
	location := fmt.Sprintf("%s:%d", filepath.Base(file), line) // 获取文件名
	return function, location
}

func GoBehind(f func()) {
	config := getConfig()
	config.Pool.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 64<<10)
				buf = buf[:runtime.Stack(buf, false)]
				log := time.Now().Format("2006-01-02 15:04:05.000")
				log += fmt.Sprintf("\t%v\n", r)
				log += " -- stack:" + string(buf)
				function, location := caller()
				config.Logger.WithField("function", function).
					WithField("location", location).
					WithField("namespace", config.NamespaceMode).
					WithField("stack", string(buf)).Error("go behind panic")
				alarm_notify.AlarmByEnv(os.Getenv("APP_MODE"), "go behind panic", log)
			}
		}()
		f()
	})
}

// GoBehindWithParam param 为可变参数 参数顺序各业务端自行控制
func GoBehindWithParam(f func([]interface{}), param ...interface{}) {
	config := getConfig()
	config.Pool.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 64<<10)
				buf = buf[:runtime.Stack(buf, false)]
				log := time.Now().Format("2006-01-02 15:04:05.000")
				log += fmt.Sprintf("\t%v\n", r)
				log += " -- stack:" + string(buf)
				function, location := caller()
				config.Logger.WithField("function", function).
					WithField("location", location).
					WithField("namespace", config.NamespaceMode).
					WithField("stack", string(buf)).Error("go behind panic")
				alarm_notify.AlarmByEnv(os.Getenv("APP_MODE"), "go behind panic", log)
			}
		}()
		f(param)
	})
}
