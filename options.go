package clog

import (
	"io"
	"os"
	"time"
)

var Set = new(setting)

type setting struct{}
type field struct{}

func NewOption() *options {
	return new(options)
}

type options struct {
	w     LevelWriter
	level Level
	// 回调函数
	prefix   []byte
	hooks    []Hook
	preHooks []Hook
}

// WithHook 添加Hook函数
func (o *options) WithHook(hook ...Hook) *options {
	o.hooks = append(o.hooks, hook...)
	return o
}

// WithPreHook 添加前置Hook函数,
func (o *options) WithPreHook(hook ...Hook) *options {
	o.preHooks = append(o.hooks, hook...)
	return o
}

// WithWriter 为日志设置输出源
func (o *options) WithWriter(w io.Writer) *options {
	if w == nil {
		w = os.Stderr
	}
	lw, ok := w.(LevelWriter)
	if !ok {
		lw = levelWriterAdapter{w}
	}
	o.w = lw
	return o
}

// WithLogLevel 设置日志等级
func (o *options) WithLogLevel(lvl Level) *options {
	o.level = lvl
	return o
}

// WithTimestamp 添加前置TimestampHook函数
func (o *options) WithTimestamp() *options {
	o.preHooks = append(o.preHooks, stp)
	return o
}

// Default 生成全局实例 调用clog.Log*前需初始化全局实例
func (o *options) Default() {
	if clog != nil {
		panic("The default function can only be called once")
	}
	clog = new(Logger)
	clog.hooks = append(clog.hooks, o.hooks...)
	clog.preHook = append(clog.preHook, o.preHooks...)
	clog.w = o.w
	clog.level = o.level
}

func (o *options) Logger() Logger {
	log := Logger{}
	log.hooks = append(log.hooks, o.hooks...)
	log.preHook = append(log.preHook, o.preHooks...)
	log.w = o.w
	log.level = o.level
	return log
}

func (s setting) TimeFormat(layout string) setting {
	if len(layout) == 0 {
		return s
	}
	timeLayoutFormat = layout
	return s
}
func (s setting) ErrHandler(f func(err error)) setting {
	errorHandler = f
	return s
}
func (s setting) ErrMarshalHandler(f func(err error) interface{}) setting {
	if f == nil {
		return s
	}
	errorMarshalFunc = f
	return s
}
func (s setting) LevelFieldToString(f func(l Level) string) setting {
	if f == nil {
		return s
	}
	levelFieldMarshalFunc = f
	return s

}
func (s setting) TimestampFunc(f func() time.Time) setting {
	if f == nil {
		return s
	}
	timestampFunc = f
	return s
}
func (s setting) ErrStackMarshal(f func(err error) interface{}) setting {
	errorStackMarshal = f
	return s
}
func (s setting) BaseTimeDurationUnit(d time.Duration) setting {
	if d == 0 {
		return s
	}
	durationFieldUnit = d
	return s
}
func (s setting) BaseTimeDurationInteger() setting {
	durationFieldInteger = true
	return s
}
func (s setting) CallMarshalFunc(f func(file string, line int) string) setting {
	if f == nil {
		return s
	}
	callerMarshalFunc = f
	return s
}
func (s setting) CallerSkipFrameCount(n int) setting {
	callerSkipFrameCount = n
	return s
}

func (s setting) FiledName() field {
	return field{}
}
func (f field) LevelFieldName(field string) field {
	if len(field) == 0 {
		return f
	}
	levelFieldName = field
	return f
}
func (f field) TimestampFieldName(field string) field {
	if len(field) == 0 {
		return f
	}
	timestampFieldName = field
	return f
}
func (f field) ErrorFieldName(field string) field {
	if len(field) == 0 {
		return f
	}
	errorFieldName = field
	return f
}
func (f field) ErrStackFieldName(field string) field {
	if len(field) == 0 {
		return f
	}
	errorStackFieldName = field
	return f
}
func (f field) MessageFieldName(field string) field {
	if len(field) == 0 {
		return f
	}
	messageFieldName = field
	return f
}
func (f field) CallerFieldName(field string) field {
	if len(field) == 0 {
		return f
	}
	callerFieldName = field
	return f
}
