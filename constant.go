package clog

import (
	"strconv"
	"sync/atomic"
	"time"
)

const (
	maxCap = 1 << 16 // 64KiB
	initCap = 1 << 9 // 512B
)

var (
	// TimeFieldFormat defines the time format of the Time field type. If set to
	// TimeFormatUnix, TimeFormatUnixMs or TimeFormatUnixMicro, the time is formatted as an UNIX
	// timestamp as integer.
	TimeFieldFormat = time.RFC3339
	// LevelFieldName is the field name used for the level field.
	LevelFieldName = "level"
	// TimestampFieldName is the field name used for the timestamp field.
	TimestampFieldName = "time"
	// ErrorFieldName is the field name used for error fields.
	ErrorFieldName = "error"
	// ErrorStackFieldName is the field name used for error stacks.
	ErrorStackFieldName = "stack"
	MessageFieldName = "message"

	// ErrorHandler 当向输出源写入数据时遇到错误，此方法调用，默认输出至stderr.
	// 此方法必须为非阻塞且线程安全
	ErrorHandler func(err error)
	ErrorMarshalFunc = func(err error) interface{} {
		return err
	}
	LevelFieldMarshalFunc = func(l Level) string {
		return l.String()
	}
	// TimestampFunc defines the function called to generate a timestamp.
	TimestampFunc = time.Now
	// ErrorStackMarshaler extract the stack from err if any.
	ErrorStackMarshaler func(err error) interface{}

	// DurationFieldUnit defines the unit for time.Duration type fields added
	// using the Dur method.
	DurationFieldUnit = time.Millisecond

	// DurationFieldInteger renders Dur fields as integer instead of float if
	// set to true.
	DurationFieldInteger = false
	// CallerFieldName is the field name used for caller field.
	CallerFieldName = "caller"

	CallerMarshalFunc = func(file string, line int) string {
		return file + ":" + strconv.Itoa(line)
	}

	// CallerSkipFrameCount is the number of stack frames to skip to find the caller.
	CallerSkipFrameCount = 2
)
var (
	gLevel          = new(int32)
	disableSampling = new(int32)
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

func samplingDisabled() bool {
	return atomic.LoadInt32(disableSampling) == 1
}