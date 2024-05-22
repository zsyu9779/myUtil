// Copyright 2014 beego Author. All Rights Reserved.
//logger 日志包
//	实现了php中logCollect， logError函数
// 	服务启动的时候调用initlog 初始化目录路径，包含大数据日志路径，业务日志路径，资源路径
// 		大数据日志路径path + log_collect
// 		业务日志路径path + log_msg
//  	资源日志路名 path + log_resource

package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"codeup.aliyun.com/aha/social_aha_gotool/uniqid"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"codeup.aliyun.com/aha/social_aha_gotool/tool"
	"codeup.aliyun.com/aha/social_aha_gotool/utils"
)

// BUFFER_CAP 异步日志队列大小
const BUFFER_CAP = 1024 * 32

// TypeResourceLog  日志类型 资源访问日志
const TypeResourceLog = "type_resource_log"

var defaultLogDir = "/tmp/golog/"
var logResourcePath string

//	localHostName 本机名称
var localHostName string

//	InitLogPath 服务启动的时候初始化业务日志路径
// 		logCollectPath string  logCollect大数据日志收集路径 /home/log/log_collect
//		logMsgPath string 业务日志路径 对应php中的logError，go中对应的函数是LogMsg /home/log/log_msg
//		logResourcePath string 资源日志路径 mysql redis http mc日志

func Init(dir, collectPathName, msgPathName, resourcePathName string) {
	if dir != "" {
		defaultLogDir = dir
	}
	InitLog(defaultLogDir+collectPathName, defaultLogDir+msgPathName, defaultLogDir+resourcePathName)

	// 初始化paniclog
	InitPanicRecoverLogger(dir + "panic.log")
}

//	InitLog 服务启动的时候初始化业务日志
// 		logCollectPath string  logCollect大数据日志收集路径 /home/log/log_collect
//		logMsgPath string 业务日志路径 对应php中的logError，go中对应的函数是LogMsg /home/log/log_msg
//		logResourcePath string 资源日志路径 mysql redis http mc日志

func InitLog(collectPath, msgPath, resourcePath string) {

	// 创建logCollect日志路径
	if collectPath == "" {
		collectPath = defaultLogDir + "log_collect/"
	}
	logCollectPath = collectPath
	if !tool.FileExists(logCollectPath) {
		if err := createFile(logCollectPath); err != nil {
			panic(err)
		}
	}

	// 创建业务日志路径
	if msgPath == "" {
		msgPath = defaultLogDir + "log_msg/"
	}
	logMsgPath = msgPath
	if !tool.FileExists(logMsgPath) {
		if err := createFile(logMsgPath); err != nil {
			panic(err)
		}
	}

	// 创建资源 mysql redis http mc日志
	if resourcePath == "" {
		resourcePath = defaultLogDir + "log_resource/"
	}
	logResourcePath = resourcePath
	if !tool.FileExists(logResourcePath) {
		if err := createFile(logResourcePath); err != nil {
			panic(err)
		}
	}
}

//	大数据日志收集
var logCollectPath string
var logCollectMutex sync.Mutex
var logCollectMap map[string]*Logger

// LogCollect 大数据日志收集方法
// 保持和 php LogCollect 一样的收集形式
func LogCollect(ctx context.Context, dbname, tablename string, fields map[string]interface{}) {
	fileName := filepath.Join("auto"+dbname, tablename)

	log, ok := logCollectMap[fileName]
	if !ok {
		logCollectMutex.Lock()
		log = getLogCollect(fileName)
		logCollectMutex.Unlock()
	}
	for k, v := range fields {
		if _, ok := delKey[k]; ok {
			delete(fields, k)
		}
		if _, ok := strvalKey[k]; ok {
			fields[k] = fmt.Sprintf("%s", v)
		}
	}
	fields["type"] = tablename
	logJsonMsg(ctx, log, fields)
}

func getLogCollect(fileName string) *Logger {
	if log, ok := logCollectMap[fileName]; ok {
		return log
	}
	filePath := filepath.Join(logCollectPath, fileName)
	log := NewLogger(filePath, "", BUFFER_CAP)
	newLogCollect := make(map[string]*Logger)
	for k, v := range logCollectMap {
		newLogCollect[k] = v
	}
	newLogCollect[fileName] = log
	logCollectMap = newLogCollect
	return log
}

// LogMsg 业务日志收集
var logMsgPath string
var logMsgMutex sync.Mutex
var logMsgMap map[string]*Logger

// LogMsg 记录业务日志
// fileName 格式  xxx 或 xxx/xxx 或 xxx/.../xxx 结尾自动追加 _Ymd.log
// ctx 传 nil 不追加 AhaReqData
func LogMsg(ctx context.Context, fileName string, args ...interface{}) {
	log, ok := logMsgMap[fileName]
	if !ok {
		logMsgMutex.Lock()
		log = getLogMsg(fileName)
		logMsgMutex.Unlock()
	}
	logMsg(ctx, log, args)
}

// LogMsgNoCut 记录业务日志，不切割
// fileName 格式  xxx 或 xxx/xxx 或 xxx/.../xxx  结尾自动追加 .log
// ctx 传 nil 不追加 AhaReqData
func LogMsgNoCut(ctx context.Context, fileName string, args ...interface{}) {
	log, ok := logMsgMap[fileName]
	if !ok {
		logMsgMutex.Lock()
		log = getLogMsg(fileName, 0)
		logMsgMutex.Unlock()
	}
	logMsg(ctx, log, args)
}

// LogJson 记录 json 日志
// ctx 传 nil 不追加 AhaReqData
// fileName 格式  xxx 或 xxx/xxx 或 xxx/.../xxx 结尾自动追加 _Ymd.log
// fields 支持格式: map[string]interface{}(建议), []byte
func LogJson(ctx context.Context, fileName string, fields interface{}) {
	log, ok := logMsgMap[fileName]
	if !ok {
		logMsgMutex.Lock()
		log = getLogMsg(fileName)
		logMsgMutex.Unlock()
	}
	var actualFields map[string]interface{}
	switch fields.(type) {
	case map[string]interface{}:
		actualFields = fields.(map[string]interface{})
	case []byte:
		err := json.Unmarshal(fields.([]byte), &actualFields)
		if err != nil {
			return
		}
	}
	logJsonMsg(ctx, log, actualFields)
}

// LogJsonNoCut 记录 json 日志，不切割
// ctx 传 nil 不追加 AhaReqData
// fileName 格式  xxx 或 xxx/xxx. xxx/.../xxx，结尾自动追加 .log
// fields 支持格式: map[string]interface{}(建议), []byte
func LogJsonNoCut(ctx context.Context, fileName string, fields interface{}) {
	log, ok := logMsgMap[fileName]
	if !ok {
		logMsgMutex.Lock()
		log = getLogMsg(fileName, 0)
		logMsgMutex.Unlock()
	}
	var actualFields map[string]interface{}
	switch fields.(type) {
	case map[string]interface{}:
		actualFields = fields.(map[string]interface{})
	case []byte:
		err := json.Unmarshal(fields.([]byte), &actualFields)
		if err != nil {
			return
		}
	}
	logJsonMsg(ctx, log, actualFields)
}

// getLogMsg 从 map 里查询是否有现成 logger，如果没有就初始化一个
// 并发不安全，需要上一层加锁
func getLogMsg(fileName string, options ...int) *Logger {
	key := fmt.Sprintf("%v:%v", fileName, options)
	if log, ok := logMsgMap[key]; ok {
		return log
	}
	filePath := filepath.Join(logMsgPath, fileName)
	log := NewLogger(filePath, "", BUFFER_CAP, options...)
	newLogMsg := make(map[string]*Logger)
	for k, v := range logMsgMap {
		newLogMsg[k] = v
	}
	newLogMsg[key] = log
	logMsgMap = newLogMsg
	return log
}

// delKey strvalKey 在 LogCollect 中强制需要删除或转换类型的key
var delKey = map[string]bool{"_type": true, "_id": true, "_score": true, "_index": true, "@version": true, "@timestamp": true}
var strvalKey = map[string]bool{"userid": true, "workid": true, "uid": true, "songid": true, "duetid": true}

// LogInfo 写入日志
// 		typ logCollc方法会传
func LogInfo(ctx context.Context, log *Logger, typ string, format interface{}, args ...interface{}) {

	logFields := make(map[string]interface{})
	now := time.Now()
	logFields["logtime"] = now.Unix()
	//logMsg["type"] = typ
	logFields["time"] = now.Format("2006-01-02 15:04:05")
	logFields["localname"] = localHostName
	for k, v := range utils.GetReqData(ctx) {
		var nk string
		switch k {
		case "uri":
			nk = "uri"
		case "version":
			nk = "version"
		case "synid":
			nk = "synid"
		case "clientip":
			nk = "clientip"
		case "userid":
			nk = "userid"
		default:
			nk = k
		}
		logFields[nk] = v
	}

	if len(args) > 0 {
		forma, ok := format.(string)
		if ok {
			logFields["msg"] = fmt.Sprintf(forma, args...)
		} else {
			arg := make([]interface{}, 0, len(args))
			logFields["msg"] = map[string]interface{}{
				"format": format,
				"args":   append(arg, args...),
			}
		}
	} else if typ == "" { //	非大数据收集日志
		logFields["msg"] = format
	} else if typ == TypeResourceLog { //	资源访问日志
		if v, ok := format.(map[string]interface{}); ok {
			logFields["cost"] = v["cost"]
			logFields["remotehost"] = v["remotehost"]
		}
		logFields["msg"] = format
	} else { // 大数据收集日志，处理特定key
		switch format.(type) {
		case map[string]interface{}:
			for k, v := range format.(map[string]interface{}) {
				if _, ok := delKey[k]; ok {
					continue
				}
				if _, ok := strvalKey[k]; ok {
					if _, ok := v.(string); !ok {
						v = fmt.Sprintf("%v", v)
					}
				}
				logFields[k] = v
			}
		default:
			logFields["msg"] = format
		}
	}

	logStr, _ := json.Marshal(logFields)
	logStr = append(logStr, []byte("\n")...)
	log.Write(logStr)
}

// logMsg 记录业务日志
// 格式为 2006-01-02 15:04:05	arg0|+|arg1|+|...|+|argN	synID	clientIP	remoteIP
// ctx 传 nil 不追加 AhaReqData
func logMsg(ctx context.Context, log *Logger, args []interface{}) {
	logTime := time.Now().Format("2006-01-02 15:04:05")
	var builder strings.Builder
	builder.WriteString(logTime)
	builder.WriteString("\t")
	for i, arg := range args {
		if i != 0 {
			builder.WriteString("|+|")
		}
		builder.WriteString(fmt.Sprintf("%v", arg))
	}
	if ctx != nil {
		cbReq, ok := utils.GetCbReqData(ctx)
		if ok {
			builder.WriteString("\t" + cbReq.GetSynID() + "\t" + cbReq.GetClientIP() + "\t" + cbReq.GetRemoteIP())
		} else if uuid, err := uniqid.NewUUID(); err == nil {
			builder.WriteString("\t" + uuid.Get() + "\t" + "1.2.3.4" + "\t" + "4.3.2.1")
		}
	}
	builder.WriteString("\n")
	_, _ = log.Write([]byte(builder.String()))
}

// logJsonMsg 往指定 logger 里写 json 日志，并追加 AhaReqData 里的数据
// ctx 传 nil 不记录 AhaReqData
func logJsonMsg(ctx context.Context, log *Logger, fields map[string]interface{}) {
	logFields := make(map[string]interface{}, len(fields)+9)
	now := time.Now()
	logFields["logtime"] = now.Unix()
	//logMsg["type"] = typ
	logFields["logdate"] = now.Format("2006-01-02 15:04:05")
	//logFields["localname"] = localHostName
	for k, v := range fields {
		logFields[k] = v
	}
	if ctx != nil {
		for k, v := range utils.GetReqData(ctx) {
			var nk string
			switch k {
			case "uri":
				continue // 没必要记 uri
			case "version":
				nk = "version"
			case "synid":
				nk = "synid"
			case "clientip":
				nk = "clientip"
			case "userid":
				nk = "userid"
			default:
				nk = k
			}
			logFields[nk] = v
		}
	}
	logStr, _ := json.Marshal(logFields)
	logStr = append(logStr, []byte("\n")...)
	_, _ = log.Write(logStr)
}

//PanicRecoverLogger pannic日志路径
var PanicRecoverLogger *Logger

func InitPanicRecoverLogger(path string) {
	if PanicRecoverLogger == nil {
		PanicRecoverLogger = NewLogger(path, "", BUFFER_CAP)
	}
}

func LogPanicRecover(str []byte) {
	if PanicRecoverLogger == nil {
		return
	}
	PanicRecoverLogger.Write(str)
}

func AccessLogger(path string) *Logger {
	return NewLogger(path, "", BUFFER_CAP)
}

func createFile(path string) error {
	if tool.FileExists(path) {
		return nil
	}
	return os.MkdirAll(path, os.ModePerm)
}

func init() {
	podNames := os.Getenv("POD_NAME")
	nodeNames := os.Getenv("NODE_NAME")
	localHostName = podNames + "" + nodeNames
}
