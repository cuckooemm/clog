package clog

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

func (l Level) String() string {
	switch l {
	case TraceLevel:
		return "trace"
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	case PanicLevel:
		return "panic"
	case NoLevel:
		return "nil"
	}
	return ""
}

type Logger struct {
	w       LevelWriter
	level   Level
	preStr  []byte
	hooks   []Hook
	preHook []Hook
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

// SetStrPrefix set prefix string
func (l *Logger) ResetStrPrefix(key string, val interface{}) {
	l.preStr = nil
	l.preStr = trs.AppendInterface(append(trs.AppendString(l.preStr, key), ':'), val)
}

func (l *Logger) AppendStrPrefix(key string, val interface{}) {
	if len(l.preStr) > 0 {
		l.preStr = append(l.preStr, ',')
	}
	l.preStr = trs.AppendInterface(append(trs.AppendString(l.preStr, key), ':'), val)
}

// GetLevel returns the current Level of l.
func (l Logger) GetLevel() Level {
	return l.level
}

// Hook returns a logger with the h Hook.
func (l Logger) Hook(h Hook) Logger {
	l.hooks = append(l.hooks, h)
	return l
}

// Trace starts a new message with trace level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Trace() *Event {
	return l.newEvent(TraceLevel, nil)
}

// Debug starts a new message with debug level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Debug() *Event {
	return l.newEvent(DebugLevel, nil)
}

// Info starts a new message with info level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Info() *Event {
	return l.newEvent(InfoLevel, nil)
}

// Warn starts a new message with warn level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Warn() *Event {
	return l.newEvent(WarnLevel, nil)
}

// Error starts a new message with error level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Error() *Event {
	return l.newEvent(ErrorLevel, nil)
}

// Err starts a new message with error level with err as a field if not nil or
// with info level if err is nil.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Err(err error) *Event {
	if err != nil {
		return l.Error().Err(err)
	}

	return l.Info()
}

// Fatal starts a new message with fatal level. The os.Exit(1) function
// is called by the Msg method, which terminates the program immediately.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Fatal() *Event {
	return l.newEvent(FatalLevel, func(msg string) { os.Exit(1) })
}

// Panic starts a new message with panic level. The panic() function
// is called by the Msg method, which stops the ordinary flow of a goroutine.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Panic() *Event {
	return l.newEvent(PanicLevel, func(msg string) { panic(msg) })
}

// WithLevel starts a new message with level. Unlike Fatal and Panic
// methods, WithLevel does not terminate the program or stop the ordinary
// flow of a gourotine when used with their respective levels.
//
// You must call Msg on the returned event in order to send the event.
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

// Log starts a new message with no level. Setting GlobalLevel to Disabled
// will still disable events produced by this method.

func (l *Logger) Log() *Event {
	return l.newEvent(NoLevel, nil)
}

// Print sends a log event using debug level and no extra field.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Print(v ...interface{}) {
	if e := l.Debug(); e.Enabled() {
		e.Msg(fmt.Sprint(v...))
	}
}

// Printf sends a log event using debug level and no extra field.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Printf(format string, v ...interface{}) {
	if e := l.Debug(); e.Enabled() {
		e.Msg(fmt.Sprintf(format, v...))
	}
}

// Write implements the io.Writer interface. This is useful to set as a writer
// for the standard library log.
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
	if l.preStr != nil && len(l.preStr) > 0 {
		e.buf = append(e.buf, l.preStr...)
	}
	for _, hook := range l.preHook {
		hook.Run(e, level, "")
	}
	e.done = done
	e.ch = l.hooks
	if level != NoLevel {
		e.Str(levelFieldName, levelFieldMarshalFunc(level))
	}
	return e
}

// should returns true if the log event should be logged.
func (l *Logger) should(lvl Level) bool {
	if lvl < l.level || lvl < GlobalLevel() {
		return false
	}
	return true
}
