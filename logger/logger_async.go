package logger

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	logger     *log.Logger
	level      LogLevel
	logFile    *os.File
	filename   string
	prefix     string
	async      *loggerBuffer
	mx         sync.RWMutex
	cutDate    bool      //是否切割日期
	createTime time.Time //创建日期
}

type loggerBuffer struct {
	buffer       *bytes.Buffer
	bufferTicker *time.Ticker
	bufferCap    int
	bufferChan   chan []byte
}

// 日志，支持同步和异步
// 异步：有缓冲的，在超大或超时后写入文件系统。
// 同步：即时写入文件系统。
// filename: 完整的文件路径
// prefix: 日志内容的前缀，每行都会加
// bufferCap: 等于0是同步写入，>0是异步写入
// params:params[0],1和不传参数代表分割日志，0代表不分割日志。
func NewLogger(filename string, prefix string, bufferCap int, options ...int) *Logger {
	//目录是否存在
	pathSliceList := strings.Split(filename, "/")
	//不是当前目录
	if len(pathSliceList) > 1 {
		pathSliceList = pathSliceList[:len(pathSliceList)-1]
		path := strings.Join(pathSliceList, "/")
		_, err := os.Stat(path)
		//创建目录
		if err != nil && os.IsNotExist(err) {
			err := os.MkdirAll(path, os.ModePerm)
			if err != nil {
				log.Fatalln("create dir failed: " + err.Error())
			}
		}
	}
	cutDate := true
	if len(options) > 0 && options[0] == 0 {
		cutDate = false
	}
	//删除超限的日志
	removeExpireLog(filepath.Dir(filename), filename, 7)
	//加日期后缀
	createTime := time.Now()
	filenameWithDate := filename
	if cutDate {
		filenameWithDate += createTime.Format("_20060102.log")
	} else {
		filenameWithDate += ".log"
	}
	//打开文件
	logFile, err := os.OpenFile(filenameWithDate, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("open log file " + filenameWithDate + " error: " + err.Error())
	}
	logger := log.New(logFile, prefix, log.LstdFlags)
	logger.SetFlags(log.Ldate | log.Ltime)
	//定义logger
	l := &Logger{
		logger:     logger,
		level:      LOG_DEBUG,
		logFile:    logFile,
		filename:   filename,
		prefix:     prefix,
		async:      nil,
		cutDate:    cutDate,
		createTime: createTime,
	}
	if bufferCap > 0 {
		//定义loggerBuffer
		l.async = &loggerBuffer{
			buffer:       bytes.NewBuffer(make([]byte, 0, bufferCap)),
			bufferCap:    bufferCap,
			bufferTicker: time.NewTicker(time.Second * 1),
			bufferChan:   make(chan []byte, bufferCap),
		}
		go l.asyncWrite()
	}
	l.logger.SetOutput(l)
	return l
}

func (l *Logger) SetLevel(lev LogLevel) {
	l.level = lev
}

func (l *Logger) Close() bool {
	_ = l.logFile.Close()
	return true
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.WithLevelf(LOG_DEBUG, format, v...)
}

func (l *Logger) InfofWithCtx(ctx context.Context, fromat string, v ...interface{}) {
	//	format = getContextContent() + format
	l.WithLevelf(LOG_INFO, fromat, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.WithLevelf(LOG_INFO, format, v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.WithLevelf(LOG_INFO, format, v...)
}

func (l *Logger) Warningf(format string, v ...interface{}) {
	l.WithLevelf(LOG_WARNING, format, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.WithLevelf(LOG_ERROR, format, v...)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.WithLevelf(LOG_FATAL, format, v...)
	os.Exit(1)
}
func (l *Logger) Fatal(v ...interface{}) {
	l.WithLevelf(LOG_FATAL, fmt.Sprint(v...))
	os.Exit(1)
}

func (l *Logger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.WithLevelf(LOG_FATAL, s)
	panic(s)
}
func (l *Logger) Println(v ...interface{}) {
	l.Infof(fmt.Sprintln(v...))
}

func (l *Logger) WithLevelf(lev LogLevel, format string, v ...interface{}) {
	if lev < l.level {
		return
	}
	format = "[" + lev.String() + "]" + " " + format
	l.Logf(format, v...)
}

func (l *Logger) Logf(format string, v ...interface{}) {
	if len(v) > 0 {
		l.logger.Printf(format, v...)
	} else {
		_ = l.logger.Output(2, format)
	}
}

//写入日志。这里实现了log.Writer
func (l *Logger) Write(p []byte) (int, error) {
	//异步写
	if l.async != nil {
		l.async.bufferChan <- p
		return len(p), nil
	}
	//同步写
	if l.cutDate {
		l.logrotate()
	}
	n, err := l.logFile.Write(p)
	if err == nil {
		err = l.logFile.Sync()
	}
	return n, err
}

//异步写盘时 - 定时器或超长时从内存写入磁盘
func (l *Logger) asyncWrite() bool {
	for {
		select {
		case <-l.async.bufferTicker.C:
			if l.async.buffer.Len() > 0 {
				l.Flush()
			}
		case record := <-l.async.bufferChan:
			l.async.buffer.Write(record)
			if l.async.buffer.Len() >= l.async.bufferCap {
				l.Flush()
			}
		}
	}
}

//异步写盘时 - 强制刷磁盘
func (l *Logger) Flush() {
	if l.cutDate {
		l.logrotate()
	}
	//刷内容从buffer到文件
	_, _ = l.logFile.Write(l.async.buffer.Bytes())
	l.async.buffer.Reset()
}

// 日志切割
// 一天一个文件
// 判断时间这步是非常耗时的。从Write()挪到Flush()QPS会从16万增加到26万
func (l *Logger) logrotate() {
	//一天一个文件
	if l.createTime.Format("2006-01-02") != time.Now().Format("2006-01-02") {
		l.mx.Lock()
		defer l.mx.Unlock()
		//加日期后缀
		createTime := time.Now()
		filenameWithDate := l.filename + "." + createTime.Format("2006-01-02")
		//文件不存在
		if exists, err := fileExists(filenameWithDate); err == nil && exists {
			return
		}
		//打开文件
		logFile, err := os.OpenFile(filenameWithDate, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln("open log file " + filenameWithDate + " error: " + err.Error())
			return
		}
		//关闭老的file
		go delayCloseFile(l.filename, l.logFile)
		//更换新的文件句柄
		l.logFile = logFile
		l.createTime = createTime
	}
}

//删除N天前的log
func removeExpireLog(path, fileNamePrefix string, day int64) {
	_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}
		if !strings.Contains(path, fileNamePrefix) {
			return nil
		}
		if time.Now().Sub(info.ModTime()) > time.Duration(day)*86400*time.Second {
			_ = os.Remove(path)
		}
		return nil
	})
}

// 延迟关闭文件句柄
func delayCloseFile(filename string, f *os.File) {
	time.Sleep(60 * time.Second) //延时60秒，担心有并发写日志，关闭文件后其他协程无法写入日志了
	err := f.Close()
	if err != nil {
		log.Fatalln("close log file " + f.Name() + " error: " + err.Error())
	}
	//删除超限的日志
	removeExpireLog(filepath.Dir(filename), filename, 7)
}

// 判断所给路径文件/文件夹是否存在
func fileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil { //os.Stat获取文件信息
		return true, nil
	} else {
		if os.IsNotExist(err) {
			return false, err
		} else {
			//文件存在，但是没有访问访问权限
			return true, err
		}
	}
}
