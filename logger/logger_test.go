package logger

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestLogInfo(t *testing.T) {
	filePath := "/tmp/log/testloginfo.log"
	fileName := "testloginfo"
	log := NewLogger(filePath, "", BUFFER_CAP)
	ctx := context.Background()
	LogInfo(ctx, log, fileName, "test test  %s", "111111")

	msg := map[string]interface{}{"a_interface": "aaa", "b_interface": "bbbb"}
	LogInfo(ctx, log, fileName, msg)

	msg1 := map[string]string{"a_string": "a1", "b_string": "b2"}
	LogInfo(ctx, log, fileName, msg1)
	// {"logtime":1640337687,"msg":{"a_string":"a1","b_string":"b2"},"time":"2021-12-24 17:21:27","type":"testloginfo"}

	msg3 := map[string]int{"a_int": 1, "b_int": 2}
	LogInfo(ctx, log, fileName, msg3)
	//{"logtime":1640337687,"msg":{"a_int":1,"b_int":2},"time":"2021-12-24 17:21:27","type":"testloginfo"}

	msg4 := map[string]interface{}{"workid": 234234, "songid": 2342.2342, "sfa": "adfaf", "duetid": []string{"a", "c"}}
	LogInfo(ctx, log, fileName, msg4)
	//{"duetid":"[a c]","logtime":1640337687,"sfa":"adfaf","songid":"2342.2342","time":"2021-12-24 17:21:27","type":"testloginfo","workid":"234234"}

	msg5 := map[string]interface{}{"_type": 234234, "del": "afaf"}
	LogInfo(ctx, log, fileName, msg5)
	//{"del":"afaf","logtime":1640337687,"time":"2021-12-24 17:21:27","type":"testloginfo"}

	msg6 := map[string]interface{}{"_type": 234234, "del": "afaf"}
	LogInfo(ctx, log, "", msg6)
	// {"logtime":1640338536,"msg":{"_type":234234,"del":"afaf"},"time":"2021-12-24 17:35:36","type":""}
}

func fakeCtx() context.Context {
	ctx := context.Background()
	cbReq := map[string]string{}
	return context.WithValue(ctx, "AhaReqData", &cbReq)
}

func TestLogMsg(t *testing.T) {
	filename := "/tmp/log/logmsg"
	ctx := fakeCtx()
	LogMsg(ctx, filename, "log with cbReqData", "hello world", "hello")
	LogMsg(nil, filename, "log no cbReqData", "hello world", "hello")
	time.Sleep(1 * time.Second)
}

func TestLogMsgNoCut(t *testing.T) {
	filename := "/tmp/log/logmsg_nocut"
	ctx := fakeCtx()
	LogMsgNoCut(ctx, filename, "log with cbReqData", "hello world", "hello")
	LogMsgNoCut(nil, filename, "log no cbReqData", "hello world", "hello")
	time.Sleep(1 * time.Second)
}

func TestLogJson(t *testing.T) {
	filename := "/tmp/log/logjson"
	ctx := fakeCtx()
	fields := map[string]interface{}{
		"field1": "json with cbReqData",
		"field2": "hello world",
		"field3": "hello",
	}
	LogJson(ctx, filename, fields)
	LogJson(nil, filename, fields)
	time.Sleep(1 * time.Second)
}

func TestLogJsonNoCut(t *testing.T) {
	filename := "/tmp/log/logjson_nocut.log"
	ctx := fakeCtx()
	fields := map[string]interface{}{
		"field1": "json with cbReqData",
		"field2": "hello world",
		"field3": "hello",
	}
	bytes, _ := json.Marshal(fields)
	LogJsonNoCut(ctx, filename, fields)
	LogJsonNoCut(nil, filename, fields)
	LogJsonNoCut(ctx, filename, bytes)
	LogJsonNoCut(nil, filename, bytes)
	time.Sleep(1 * time.Second)

}
