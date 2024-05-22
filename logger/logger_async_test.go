package logger

import (
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	log := NewLogger("/var/www/golang/src/gotool/logger/test.log", "", 0)
	log.SetLevel(LOG_DEBUG)
	log.Debugf("测试%s", "Debugf")
	log.Infof("测试%s", "Infof")
	log.Warningf("测试%s", "Warningf")
	log.Errorf("测试%s", "Errorf")
	log.Fatalf("测试%s", "Fatalf")
}

func TestNewLoggerAsync(t *testing.T) {
	log := NewLogger("/var/www/golang/src/gotool/logger/test.log", "[ASYNC]", 10*1024)
	log.Logf("测试%s", "buffer")
	time.Sleep(2 * time.Second)
}
