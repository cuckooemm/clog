package clog

import (
	"io"
	"io/ioutil"
	"time"
)

var Set = Setting{}

type Options struct {
	w     LevelWriter
	level Level
	// 回调函数
	prefix   []byte
	hooks    []Hook
	preHooks []Hook
}

// 添加 Hook
func (o *Options) WithHook(hook ...Hook) *Options {
	o.hooks = append(o.hooks, hook...)
	return o
}

func (o *Options) WithWriter(w io.Writer) *Options {
	if w == nil {
		w = ioutil.Discard
	}
	lw, ok := w.(LevelWriter)
	if !ok {
		lw = levelWriterAdapter{w}
	}
	o.w = lw
	return o
}

func (o *Options) WithLogLevel(lvl Level) *Options {
	o.level = lvl
	return o
}

func (o *Options) WithPrefix(prefix string) *Options {
	o.prefix = []byte(prefix)
	return o
}

func (o *Options) WithTimestamp() *Options {
	o.preHooks = append(o.hooks, th)
	return o
}

func (o *Options) Default() {
	if clog != nil {
		panic("The default function can only be called once")
	}
	clog = new(Logger)
	clog.hooks = append(clog.hooks, o.hooks...)
	clog.preHook = append(clog.preHook, o.preHooks...)
	clog.w = o.w
	clog.level = o.level
	clog.context = append(clog.context, o.prefix...)
}

func (o *Options) Logger() Logger {
	log := Logger{}
	log.hooks = append(clog.hooks, o.hooks...)
	clog.preHook = append(clog.preHook, o.preHooks...)
	clog.w = o.w
	clog.level = o.level
	clog.context = append(clog.context, o.prefix...)
	return log
}

type Setting struct{}
type Field struct{}

func (s Setting) SetTimeFormat(layout string) Setting {
	timeLayoutFormat = layout
	return s
}
func (s Setting) SetErrHandler(f func(err error)) Setting {
	errorHandler = f
	return s
}
func (s Setting) SetErrMarshalHandler(f func(err error) interface{}) Setting {
	errorMarshalFunc = f
	return s
}
func (s Setting) SetLevelFieldToString(f func(l Level) string) Setting {
	levelFieldMarshalFunc = f
	return s

}
func (s Setting) SetTimestampFunc(f func() time.Time) Setting {
	timestampFunc = f
	return s
}
func (s Setting) SetErrStackMarshaler(f func(err error) interface{}) Setting {
	errorStackMarshaler = f
	return s
}
func (s Setting) SetBaseTimeDurationUnit(d time.Duration) Setting {
	durationFieldUnit = d
	return s
}
func (s Setting) SetBaseTimeDurationInteger(b bool) Setting {
	durationFieldInteger = b
	return s
}
func (s Setting) SetCallMarshalFunc(f func(file string, line int) string) Setting {
	callerMarshalFunc = f
	return s
}
func (s Setting) SetCallerSkipFrameCount(n int) Setting {
	callerSkipFrameCount = n
	return s
}

func (s Setting) SetFiledName() Field {
	return Field{}
}
func (f Field) SetLevelFieldName(field string) Field {
	levelFieldName = field
	return f
}
func (f Field) SetTimestampFieldName(field string) Field {
	timestampFieldName = field
	return f
}
func (f Field) SetErrorFieldName(field string) Field {
	errorFieldName = field
	return f
}
func (f Field) SetErrStackFieldName(field string) Field {
	errorStackFieldName = field
	return f
}
func (f Field) SetMessageFieldName(field string) Field {
	messageFieldName = field
	return f
}
func (f Field) SetCallerFieldName(field string) Field {
	callerFieldName = field
	return f
}
