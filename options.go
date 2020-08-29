package clog

import (
	"io"
	"io/ioutil"
	"time"
)

var (
	Set = setting{}
)

type setting struct{}
type field struct{}

func NewOption() *options {
	return &options{}
}

type options struct {
	w     LevelWriter
	level Level
	// 回调函数
	prefix   []byte
	hooks    []Hook
	preHooks []Hook
}

// 添加 Hook
func (o *options) WithHook(hook ...Hook) *options {
	o.hooks = append(o.hooks, hook...)
	return o
}

func (o *options) WithPreHook(hook ...Hook) *options {
	o.preHooks = append(o.hooks, hook...)
	return o
}

func (o *options) WithWriter(w io.Writer) *options {
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

func (o *options) WithLogLevel(lvl Level) *options {
	o.level = lvl
	return o
}

func (o *options) WithPrefix(key, value string) *options {
	o.prefix = trs.AppendString(append(trs.AppendString(o.prefix, key), ':'), value)
	return o
}

func (o *options) WithTimestamp() *options {
	o.preHooks = append(o.hooks, th)
	return o
}

func (o *options) Default() {
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

func (o *options) Logger() Logger {
	log := Logger{}
	log.hooks = append(clog.hooks, o.hooks...)
	clog.preHook = append(clog.preHook, o.preHooks...)
	clog.w = o.w
	clog.level = o.level
	clog.context = append(clog.context, o.prefix...)
	return log
}

func (s setting) SetTimeFormat(layout string) setting {
	timeLayoutFormat = layout
	return s
}
func (s setting) SetErrHandler(f func(err error)) setting {
	errorHandler = f
	return s
}
func (s setting) SetErrMarshalHandler(f func(err error) interface{}) setting {
	errorMarshalFunc = f
	return s
}
func (s setting) SetLevelFieldToString(f func(l Level) string) setting {
	levelFieldMarshalFunc = f
	return s

}
func (s setting) SetTimestampFunc(f func() time.Time) setting {
	timestampFunc = f
	return s
}
func (s setting) SetErrStackMarshaler(f func(err error) interface{}) setting {
	errorStackMarshaler = f
	return s
}
func (s setting) SetBaseTimeDurationUnit(d time.Duration) setting {
	durationFieldUnit = d
	return s
}
func (s setting) SetBaseTimeDurationInteger() setting {
	durationFieldInteger = true
	return s
}
func (s setting) SetCallMarshalFunc(f func(file string, line int) string) setting {
	callerMarshalFunc = f
	return s
}
func (s setting) SetCallerSkipFrameCount(n int) setting {
	callerSkipFrameCount = n
	return s
}

func (s setting) SetFiledName() field {
	return field{}
}
func (f field) SetLevelFieldName(field string) field {
	levelFieldName = field
	return f
}
func (f field) SetTimestampFieldName(field string) field {
	timestampFieldName = field
	return f
}
func (f field) SetErrorFieldName(field string) field {
	errorFieldName = field
	return f
}
func (f field) SetErrStackFieldName(field string) field {
	errorStackFieldName = field
	return f
}
func (f field) SetMessageFieldName(field string) field {
	messageFieldName = field
	return f
}
func (f field) SetCallerFieldName(field string) field {
	callerFieldName = field
	return f
}
