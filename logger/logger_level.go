package logger

import (
	"strings"
)

type LogLevel int

const (
	LOG_DEBUG LogLevel = iota
	LOG_INFO
	LOG_WARNING
	LOG_ERROR
	LOG_FATAL
)

func (l LogLevel) String() string {
	switch l {
	case LOG_DEBUG:
		return "debug"
	case LOG_INFO:
		return "info"
	case LOG_WARNING:
		return "warning"
	case LOG_ERROR:
		return "error"
	case LOG_FATAL:
		return "fatal"
	default:
		return "info"
	}
}

func getLevelByString(levelString string) LogLevel {
	switch strings.ToLower(levelString) {
	case "debug":
		return LOG_DEBUG
	case "info":
		return LOG_INFO
	case "warning":
		return LOG_WARNING
	case "error":
		return LOG_ERROR
	case "fatal":
		return LOG_FATAL
	default:
		return LOG_INFO
	}
}
