package clog

import (
	"strconv"
	"sync/atomic"
	"time"
)

const (
	maxCap     = 1 << 14 // 16KiB
	initCap    = 1 << 9  // 512B
	RandomLine = int64(10000)
)

// Level defines log levels.
type Level int8

const (
	// DebugLevel debug 日志等级.
	DebugLevel Level = iota
	// RandomLevel random 日志等级
	RandomLevel
	// InfoLevel info 日志等级.
	InfoLevel
	// WarnLevel warn 日志等级.
	WarnLevel
	// ErrorLevel error 日志等级.
	ErrorLevel
	// FatalLevel fatal 日志等级.
	FatalLevel
	// PanicLevel panic 日志等级.
	PanicLevel
	// NoLevel 缺省日志级别.
	NoLevel
	// Disabled 关闭log.
	Disabled

	// TraceLevel trace 日志等级.
	TraceLevel Level = -1
)

func init() {
	*gLevel = -1
}

var (
	// 全局日志等级
	gLevel = new(int32)

	// timeLayoutFormat 定义time字段日期格式
	timeLayoutFormat = time.RFC3339

	// levelFieldName level字段输出key
	levelFieldName = "level"

	// timestampFieldName Timestamp() 方法日期key
	timestampFieldName = "time"
	// errorFieldName 错误信息字段key Err()方法key
	errorFieldName = "error"
	// errorStackFieldName 栈信息key
	errorStackFieldName = "stack"

	// messageFieldName Msg()方法key
	messageFieldName = "message"
	// callerFieldName caller()方法key
	callerFieldName = "caller"

	// errorHandler 当向输出源写入数据时遇到错误，此方法调用，默认输出至stderr. 此方法必须为非阻塞且线程安全
	errorHandler     func(err error)
	errorMarshalFunc = func(err error) interface{} {
		return err
	}
	levelFieldMarshalFunc = func(l Level) string {
		return l.String()
	}
	// timestampFunc 调用函数时生成时间的方法
	timestampFunc = time.Now
	// errorStackMarshal extract the stack from err if any.
	errorStackMarshal func(err error) interface{}
	// durationFieldUnit defines the unit for time.Duration type fields added
	// timeDuration / durationFieldUnit
	durationFieldUnit = time.Millisecond

	// durationFieldInteger 为true则以整形输出时间戳
	durationFieldInteger = false

	callerMarshalFunc = func(file string, line int) string {
		return file + ":" + strconv.Itoa(line)
	}

	// callerSkipFrameCount is the number of stack frames to skip to find the caller.
	callerSkipFrameCount = 2
)

// SetGlobalLevel 设置全局日志等级,通过此全局变量控制全局日志输出.
func SetGlobalLevel(l Level) {
	atomic.StoreInt32(gLevel, int32(l))
}

// GlobalLevel 返回当前全局日志等级.
func GlobalLevel() Level {
	return Level(atomic.LoadInt32(gLevel))
}
