package clog

import (
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	maxCap  = 1 << 16 // 64KiB
	initCap = 1 << 9  // 512B
)

// Level defines log levels.
type Level int8

const (
	// DebugLevel defines debug log level.
	DebugLevel Level = iota
	// InfoLevel defines info log level.
	InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel
	// PanicLevel defines panic log level.
	PanicLevel
	// NoLevel defines an absent log level.
	NoLevel
	// Disabled disables the logger.
	Disabled

	// TraceLevel defines trace log level.
	TraceLevel Level = -1
)

func init() {
	curPath, _ := os.Getwd()
	curPathIdx = len(curPath)
}

var (
	// 全局日志等级
	gLevel     = new(int32)
	curPathIdx int
	// TimeFieldFormat defines the time format of the Time field type. If set to
	// TimeFormatUnix, TimeFormatUnixMs or TimeFormatUnixMicro, the time is formatted as an UNIX
	// timestamp as integer.
	timeLayoutFormat = time.RFC3339
	// LevelFieldName is the field name used for the level field.
	levelFieldName = "level"
	// TimestampFieldName is the field name used for the timestamp field.
	timestampFieldName = "time"
	// ErrorFieldName is the field name used for error fields.
	errorFieldName = "error"
	// ErrorStackFieldName is the field name used for error stacks.
	errorStackFieldName = "stack"
	messageFieldName    = "message"
	// CallerFieldName is the field name used for caller field.
	callerFieldName = "caller"
	// ErrorHandler 当向输出源写入数据时遇到错误，此方法调用，默认输出至stderr.
	// 此方法必须为非阻塞且线程安全
	errorHandler     func(err error)
	errorMarshalFunc = func(err error) interface{} {
		return err
	}
	levelFieldMarshalFunc = func(l Level) string {
		return l.String()
	}
	// timestampFunc defines the function called to generate a timestamp.
	timestampFunc = time.Now
	// ErrorStackMarshaler extract the stack from err if any.
	errorStackMarshaler func(err error) interface{}

	// durationFieldUnit defines the unit for time.Duration type fields added
	// timeDuration / durationFieldUnit
	durationFieldUnit = time.Millisecond

	// DurationFieldInteger renders Dur fields as integer instead of float if
	// set to true.
	durationFieldInteger = false

	callerMarshalFunc = func(file string, line int) string {
		if curPathIdx > 0 {
			return "." + file[curPathIdx:] + ":" + strconv.Itoa(line)
		}
		return file + ":" + strconv.Itoa(line)
	}

	// CallerSkipFrameCount is the number of stack frames to skip to find the caller.
	callerSkipFrameCount = 2
)

// SetGlobalLevel sets the global override for log level. If this
// values is raised, all Loggers will use at least this value.
//
// To globally disable logs, set GlobalLevel to Disabled.
func SetGlobalLevel(l Level) {
	atomic.StoreInt32(gLevel, int32(l))
}

// GlobalLevel returns the current global log level
func GlobalLevel() Level {
	return Level(atomic.LoadInt32(gLevel))
}
