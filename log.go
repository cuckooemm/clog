package clog

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

type Logger struct {
	w       LevelWriter
	level   Level
	preHook []Hook
	hooks   []Hook
}

func ParseLevel(levelStr string) (Level, error) {
	switch levelStr {
	case levelFieldMarshalFunc(TraceLevel):
		return TraceLevel, nil
	case levelFieldMarshalFunc(DebugLevel):
		return DebugLevel, nil
	case levelFieldMarshalFunc(InfoLevel):
		return InfoLevel, nil
	case levelFieldMarshalFunc(WarnLevel):
		return WarnLevel, nil
	case levelFieldMarshalFunc(ErrorLevel):
		return ErrorLevel, nil
	case levelFieldMarshalFunc(NoLevel):
		return NoLevel, nil
	case levelFieldMarshalFunc(Disabled):
		return Disabled, nil
	}
	return NoLevel, fmt.Errorf("unknown Level String: '%s', defaulting to NoLevel", levelStr)
}

// Output duplicates the current logger and sets w as its output.
func (l *Logger) output(w io.Writer) {
	lw, ok := w.(LevelWriter)
	if !ok {
		lw = levelWriterAdapter{w}
	}
	l.w = lw
}

// GetLevel 返回当前实例的日志等级.
func (l Logger) GetLevel() Level {
	return l.level
}

// Hook 向事件添加hook.
func (l *Logger) Hook(h Hook) *Logger {
	l.hooks = append(l.hooks, h)
	return l
}

// Trace 开启一个trace等级的日志事件.
// 必须调用Msg()方法完成此事件.
func (l *Logger) Trace() *Event {
	return l.newEvent(TraceLevel, nil)
}

// Debug 开启一个debug等级的日志事件.
// 必须调用Msg()方法完成此事件.
func (l *Logger) Debug() *Event {
	return l.newEvent(DebugLevel, nil)
}

// Info 开启一个info等级的日志事件.
// 必须调用Msg()方法完成此事件.
func (l *Logger) Info() *Event {
	return l.newEvent(InfoLevel, nil)
}

// Warn 开启一个warn等级的日志事件.
// 必须调用Msg()方法完成此事件.
func (l *Logger) Warn() *Event {
	return l.newEvent(WarnLevel, nil)
}

// Error 开启一个error等级的日志事件.
// 必须调用Msg()方法完成此事件.
func (l *Logger) Error() *Event {
	return l.newEvent(ErrorLevel, nil)
}

// Err 根据传入的error判断开启的事件等级,如果err不为空,则设置事件等级为error，否则为info.
// 必须调用Msg()方法完成此事件.
func (l *Logger) Err(err error) *Event {
	if err != nil {
		return l.Error().Err(err)
	}

	return l.Info()
}

// Fatal 开启一个Fatal等级的日志事件. 调用Msg完成事件是同时调用 os.Exit(1)函数并退出程序.
// 必须调用Msg()方法完成此事件.
func (l *Logger) Fatal() *Event {
	return l.newEvent(FatalLevel, func(msg string) { os.Exit(1) })
}

// Panic 开启一个Panic等级的日志事件,且在完成事件时调用panic.
//
// 必须调用Msg()方法完成此事件.
func (l *Logger) Panic() *Event {
	return l.newEvent(PanicLevel, func(msg string) { panic(msg) })
}

// WithLevel 根据传入的等级生成时间,如果传入panic,fatal等级,不会调用panic,os.exit等相关函数.
// 必须调用Msg()方法完成此事件.
func (l *Logger) WithLevel(level Level) *Event {
	switch level {
	case TraceLevel:
		return l.Trace()
	case DebugLevel:
		return l.Debug()
	case InfoLevel:
		return l.Info()
	case WarnLevel:
		return l.Warn()
	case ErrorLevel:
		return l.Error()
	case FatalLevel:
		return l.newEvent(FatalLevel, nil)
	case PanicLevel:
		return l.newEvent(PanicLevel, nil)
	case NoLevel:
		return l.Log()
	case Disabled:
		return nil
	default:
		panic("clog: WithLevel(): invalid level: " + strconv.Itoa(int(level)))
	}
}

// Log 开启一个无等级的日志事件,输出中不会包含level等相关字段信息.
// 设置全局日志等级 GlobalLevel 为 Disabled 将禁用此方法产生的事件.
func (l *Logger) Log() *Event {
	return l.newEvent(NoLevel, nil)
}

// Print 使用debug日志等级且不包含key形式输出.
// 与 fmt.Print 参数相同.
func (l *Logger) Print(v ...interface{}) {
	if e := l.Debug(); e.IsEnabled() {
		e.Msg(fmt.Sprint(v...))
	}
}

// Printf 使用debug日志等级且不包含key形式输出.
// 与 fmt.Printf 参数相同.
func (l *Logger) Printf(format string, v ...interface{}) {
	if e := l.Debug(); e.IsEnabled() {
		e.Msgf(format, v...)
	}
}

// Write 实现 io.Writer 接口.
func (l Logger) Write(p []byte) (n int, err error) {
	n = len(p)
	if n > 0 && p[n-1] == '\n' {
		p = p[:n-1]
	}
	l.Log().Msg(string(p))
	return
}

func (l *Logger) newEvent(level Level, done func(string)) *Event {
	if !l.should(level) {
		return nil
	}
	e := newEvent(l.w, level)
	for _, hook := range l.preHook {
		hook.Run(e, level, "")
	}
	e.done = done
	e.hook = l.hooks
	if level != NoLevel {
		e.Str(levelFieldName, levelFieldMarshalFunc(level))
	}
	return e
}

// should 如果log等级小于实例等级或小于全局等级,则返回True.
func (l Logger) should(lvl Level) bool {
	if lvl < l.level || lvl < GlobalLevel() {
		return false
	}
	return true
}
