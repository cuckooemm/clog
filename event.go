package clog

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"sync"
	"time"
)

var eventPool = &sync.Pool{
	New: func() interface{} {
		return &Event{
			buf: make([]byte, 0, initCap),
		}
	},
}

// Event 代表一个日志事件，由Level方法实例化，Msg完成。
type Event struct {
	buf   []byte
	w     LevelWriter
	level Level
	done  func(msg string)
	stack bool   // enable error stack trace
	ch    []Hook // hooks from context
}

type LogObjectMarshaler interface {
	MarshalObject(e *Event)
}

// LogArrayMarshaler provides a strongly-typed and encoding-agnostic interface
// to be implemented by types used with Event/Context's Array methods.
type LogArrayMarshaler interface {
	MarshalArray(a *Array)
}

func newEvent(w LevelWriter, level Level) *Event {
	e := eventPool.Get().(*Event)
	e.buf = e.buf[:0]
	e.ch = nil
	e.buf = trs.AppendBeginMarker(e.buf)
	e.w = w
	e.level = level
	e.stack = false
	return e
}

func putEvent(e *Event) {
	if cap(e.buf) > maxCap {
		e.buf = e.buf[:0:initCap]
		//return
	}
	eventPool.Put(e)
}

func (e *Event) write() (err error) {
	if e == nil {
		return nil
	}
	if e.level != Disabled {
		e.buf = trs.AppendEndMarker(e.buf)
		e.buf = trs.AppendLineBreak(e.buf)
		if e.w != nil {
			_, err = e.w.WriteLevel(e.level, e.buf)
		}
	}
	putEvent(e)
	return
}

// Enabled return false if the *Event is going to be filtered out by
// log level or sampling.
func (e *Event) Enabled() bool {
	return e != nil && e.level != Disabled
}

// Discard disables the event so Msg(f) won't print it.
func (e *Event) Discard() *Event {
	if e == nil {
		return e
	}
	e.level = Disabled
	return nil
}

// Msg sends the *Event with msg added as the message field if not empty.
//
// NOTICE: once this method is called, the *Event should be disposed.
// Calling Msg twice can have unexpected result.
func (e *Event) Msg(msg string) {
	if e == nil {
		return
	}
	e.msg(msg)
}

// Done is equivalent to calling Msg("").
//
// NOTICE: once this method is called, the *Event should be disposed.
func (e *Event) Done() {
	if e == nil {
		return
	}
	e.msg("")
}

func (e *Event) Msgf(format string, v ...interface{}) {
	if e == nil {
		return
	}
	e.msg(fmt.Sprintf(format, v...))
}

func (e *Event) msg(msg string) {
	for _, hook := range e.ch {
		hook.Run(e, e.level, msg)
	}
	if msg != "" {
		e.buf = trs.AppendString(trs.AppendKey(e.buf, messageFieldName), msg)
	}
	if e.done != nil {
		defer e.done(msg)
	}
	if err := e.write(); err != nil {
		if errorHandler != nil {
			errorHandler(err)
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "clog: could not write to output: %v\n", err)
		}
	}
}

// Fields is a helper function to use a map to set fields using type assertion.
func (e *Event) Fields(fields map[string]interface{}) *Event {
	if e == nil {
		return e
	}
	e.buf = appendFields(e.buf, fields)
	return e
}

// Dict adds the field key with a dict to the event context.
// Use clog.Dict() to create the dictionary.
func (e *Event) Dict(key string, dict *Event) *Event {
	if e == nil {
		return e
	}
	dict.buf = trs.AppendEndMarker(dict.buf)
	e.buf = append(trs.AppendKey(e.buf, key), dict.buf...)
	putEvent(dict)
	return e
}

// Dict creates an Event to be used with the *Event.Dict method.
// Call usual field methods like Str, Int etc to add fields to this
// event and give it as argument the *Event.Dict method.
func Dict() *Event {
	return newEvent(nil, 0)
}

// Array adds the field key with an array to the event context.
// Use clog.Arr() to create the array or pass a type that
// implement the LogArrayMarshaler interface.
func (e *Event) Array(key string, arr LogArrayMarshaler) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendKey(e.buf, key)
	var a *Array
	if aa, ok := arr.(*Array); ok {
		a = aa
	} else {
		a = Arr()
		arr.MarshalArray(a)
	}
	e.buf = a.write(e.buf)
	return e
}

func (e *Event) appendObject(obj LogObjectMarshaler) {
	e.buf = trs.AppendBeginMarker(e.buf)
	obj.MarshalObject(e)
	e.buf = trs.AppendEndMarker(e.buf)
}

// Object marshals an object that implement the LogObjectMarshaler interface.
func (e *Event) Object(key string, obj LogObjectMarshaler) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendKey(e.buf, key)
	e.appendObject(obj)
	return e
}

// EmbedObject marshals an object that implement the LogObjectMarshaler interface.
func (e *Event) EmbedObject(obj LogObjectMarshaler) *Event {
	if e == nil {
		return e
	}
	obj.MarshalObject(e)
	return e
}

// Str adds the field key with val as a string to the *Event context.
func (e *Event) Str(key, val string) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendString(trs.AppendKey(e.buf, key), val)
	return e
}

// Strs adds the field key with vals as a []string to the *Event context.
func (e *Event) Strs(key string, vals []string) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendStrings(trs.AppendKey(e.buf, key), vals)
	return e
}

// Stringer adds the field key with val.String() (or null if val is nil) to the *Event context.
func (e *Event) Stringer(key string, val fmt.Stringer) *Event {
	if e == nil {
		return e
	}
	if val != nil {
		e.buf = trs.AppendString(trs.AppendKey(e.buf, key), val.String())
		return e
	}
	e.buf = trs.AppendInterface(trs.AppendKey(e.buf, key), nil)
	return e
}

// Bytes adds the field key with val as a string to the *Event context.
//
// Runes outside of normal ASCII ranges will be hex-encoded in the resulting
// JSON.
func (e *Event) Bytes(key string, val []byte) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendBytes(trs.AppendKey(e.buf, key), val)
	return e
}

// Hex adds the field key with val as a hex string to the *Event context.
func (e *Event) Hex(key string, val []byte) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendHex(trs.AppendKey(e.buf, key), val)
	return e
}

// RawJSON adds already encoded JSON to the log line under key.
//
// No sanity check is performed on b; it must not contain carriage returns and
// be valid JSON.
func (e *Event) RawJSON(key string, b []byte) *Event {
	if e == nil {
		return e
	}
	e.buf = append(trs.AppendKey(e.buf, key), b...)
	return e
}

// AnErr adds the field key with serialized err to the *Event context.
// If err is nil, no field is added.
func (e *Event) AnErr(key string, err error) *Event {
	if e == nil {
		return e
	}
	switch m := errorMarshalFunc(err).(type) {
	case nil:
		return e
	case LogObjectMarshaler:
		return e.Object(key, m)
	case error:
		if m == nil || isNilValue(m) {
			return e
		} else {
			return e.Str(key, m.Error())
		}
	case string:
		return e.Str(key, m)
	default:
		return e.Interface(key, m)
	}
}

// Errs adds the field key with errs as an array of serialized errors to the
// *Event context.
func (e *Event) Errs(key string, errs []error) *Event {
	if e == nil {
		return e
	}
	arr := Arr()
	for _, err := range errs {
		switch m := errorMarshalFunc(err).(type) {
		case LogObjectMarshaler:
			arr = arr.Object(m)
		case error:
			arr = arr.Err(m)
		case string:
			arr = arr.Str(m)
		default:
			arr = arr.Interface(m)
		}
	}

	return e.Array(key, arr)
}

// Err adds the field "error" with serialized err to the *Event context.
// If err is nil, no field is added.
//
// To customize the key name, change clog.ErrorFieldName.
//
// If Stack() has been called before and clog.ErrorStackMarshaler is defined,
// the err is passed to ErrorStackMarshaler and the result is appended to the
// clog.ErrorStackFieldName.
func (e *Event) Err(err error) *Event {
	if e == nil {
		return e
	}
	if e.stack && errorStackMarshaler != nil {
		switch m := errorStackMarshaler(err).(type) {
		case nil:
		case LogObjectMarshaler:
			e.Object(errorStackFieldName, m)
		case error:
			if m != nil && !isNilValue(m) {
				e.Str(errorStackFieldName, m.Error())
			}
		case string:
			e.Str(errorStackFieldName, m)
		default:
			e.Interface(errorStackFieldName, m)
		}
	}
	return e.AnErr(errorFieldName, err)
}

// Stack enables stack trace printing for the error passed to Err().
//
// ErrorStackMarshaler must be set for this method to do something.
func (e *Event) Stack() *Event {
	if e == nil {
		return e
	}
	e.stack = true
	return e
}

// Bool adds the field key with val as a bool to the *Event context.
func (e *Event) Bool(key string, b bool) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendBool(trs.AppendKey(e.buf, key), b)
	return e
}

// Bools adds the field key with val as a []bool to the *Event context.
func (e *Event) Bools(key string, b []bool) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendBools(trs.AppendKey(e.buf, key), b)
	return e
}

// Int adds the field key with i as a int to the *Event context.
func (e *Event) Int(key string, i int) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInt(trs.AppendKey(e.buf, key), i)
	return e
}

// Ints adds the field key with i as a []int to the *Event context.
func (e *Event) Ints(key string, i []int) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInts(trs.AppendKey(e.buf, key), i)
	return e
}

// Int8 adds the field key with i as a int8 to the *Event context.
func (e *Event) Int8(key string, i int8) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInt8(trs.AppendKey(e.buf, key), i)
	return e
}

// Ints8 adds the field key with i as a []int8 to the *Event context.
func (e *Event) Ints8(key string, i []int8) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInts8(trs.AppendKey(e.buf, key), i)
	return e
}

// Int16 adds the field key with i as a int16 to the *Event context.
func (e *Event) Int16(key string, i int16) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInt16(trs.AppendKey(e.buf, key), i)
	return e
}

// Ints16 adds the field key with i as a []int16 to the *Event context.
func (e *Event) Ints16(key string, i []int16) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInts16(trs.AppendKey(e.buf, key), i)
	return e
}

// Int32 adds the field key with i as a int32 to the *Event context.
func (e *Event) Int32(key string, i int32) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInt32(trs.AppendKey(e.buf, key), i)
	return e
}

// Ints32 adds the field key with i as a []int32 to the *Event context.
func (e *Event) Ints32(key string, i []int32) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInts32(trs.AppendKey(e.buf, key), i)
	return e
}

// Int64 adds the field key with i as a int64 to the *Event context.
func (e *Event) Int64(key string, i int64) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInt64(trs.AppendKey(e.buf, key), i)
	return e
}

// Int64s adds the field key with i as a []int64 to the *Event context.
func (e *Event) Ints64(key string, i []int64) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInts64(trs.AppendKey(e.buf, key), i)
	return e
}

// Uint adds the field key with i as a uint to the *Event context.
func (e *Event) Uint(key string, i uint) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUint(trs.AppendKey(e.buf, key), i)
	return e
}

// Uints adds the field key with i as a []int to the *Event context.
func (e *Event) Uints(key string, i []uint) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUints(trs.AppendKey(e.buf, key), i)
	return e
}

// Uint8 adds the field key with i as a uint8 to the *Event context.
func (e *Event) Uint8(key string, i uint8) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUint8(trs.AppendKey(e.buf, key), i)
	return e
}

// Uints8 adds the field key with i as a []int8 to the *Event context.
func (e *Event) Uints8(key string, i []uint8) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUints8(trs.AppendKey(e.buf, key), i)
	return e
}

// Uint16 adds the field key with i as a uint16 to the *Event context.
func (e *Event) Uint16(key string, i uint16) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUint16(trs.AppendKey(e.buf, key), i)
	return e
}

// Uints16 adds the field key with i as a []int16 to the *Event context.
func (e *Event) Uints16(key string, i []uint16) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUints16(trs.AppendKey(e.buf, key), i)
	return e
}

// Uint32 adds the field key with i as a uint32 to the *Event context.
func (e *Event) Uint32(key string, i uint32) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUint32(trs.AppendKey(e.buf, key), i)
	return e
}

// Uints32 adds the field key with i as a []int32 to the *Event context.
func (e *Event) Uints32(key string, i []uint32) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUints32(trs.AppendKey(e.buf, key), i)
	return e
}

// Uint64 adds the field key with i as a uint64 to the *Event context.
func (e *Event) Uint64(key string, i uint64) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUint64(trs.AppendKey(e.buf, key), i)
	return e
}

// Uints64 adds the field key with i as a []int64 to the *Event context.
func (e *Event) Uints64(key string, i []uint64) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUints64(trs.AppendKey(e.buf, key), i)
	return e
}

// Float32 adds the field key with f as a float32 to the *Event context.
func (e *Event) Float32(key string, f float32) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendFloat32(trs.AppendKey(e.buf, key), f)
	return e
}

// Floats32 adds the field key with f as a []float32 to the *Event context.
func (e *Event) Floats32(key string, f []float32) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendFloats32(trs.AppendKey(e.buf, key), f)
	return e
}

// Float64 adds the field key with f as a float64 to the *Event context.
func (e *Event) Float64(key string, f float64) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendFloat64(trs.AppendKey(e.buf, key), f)
	return e
}

// Floats64 adds the field key with f as a []float64 to the *Event context.
func (e *Event) Floats64(key string, f []float64) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendFloats64(trs.AppendKey(e.buf, key), f)
	return e
}

// Timestamp adds the current local time as UNIX timestamp to the *Event context with the "time" key.
// To customize the key name, change clog.TimestampFieldName.
//
// NOTE: It won't dedupe the "time" key if the *Event (or *Context) has one
// already.
func (e *Event) Timestamp() *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendTime(trs.AppendKey(e.buf, timestampFieldName), timestampFunc(), timeLayoutFormat)
	return e
}

// Time adds the field key with t formated as string using clog.TimeFieldFormat.
func (e *Event) Time(key string, t time.Time) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendTime(trs.AppendKey(e.buf, key), t, timeLayoutFormat)
	return e
}

// Times adds the field key with t formated as string using clog.TimeFieldFormat.
func (e *Event) Times(key string, t []time.Time) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendTimes(trs.AppendKey(e.buf, key), t, timeLayoutFormat)
	return e
}

// Dur adds the field key with duration d stored as clog.DurationFieldUnit.
// If clog.DurationFieldInteger is true, durations are rendered as integer
// instead of float.
func (e *Event) Dur(key string, d time.Duration) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendDuration(trs.AppendKey(e.buf, key), d, durationFieldUnit, durationFieldInteger)
	return e
}

// Durs adds the field key with duration d stored as clog.DurationFieldUnit.
// If clog.DurationFieldInteger is true, durations are rendered as integer
// instead of float.
func (e *Event) Durs(key string, d []time.Duration) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendDurations(trs.AppendKey(e.buf, key), d, durationFieldUnit, durationFieldInteger)
	return e
}

// TimeDiff adds the field key with positive duration between time t and start.
// If time t is not greater than start, duration will be 0.
// Duration format follows the same principle as Dur().
func (e *Event) TimeDiff(key string, t time.Time, start time.Time) *Event {
	if e == nil {
		return e
	}
	var d time.Duration
	if t.After(start) {
		d = t.Sub(start)
	}
	e.buf = trs.AppendDuration(trs.AppendKey(e.buf, key), d, durationFieldUnit, durationFieldInteger)
	return e
}

// Interface adds the field key with i marshaled using reflection.
func (e *Event) Interface(key string, i interface{}) *Event {
	if e == nil {
		return e
	}
	if obj, ok := i.(LogObjectMarshaler); ok {
		return e.Object(key, obj)
	}
	e.buf = trs.AppendInterface(trs.AppendKey(e.buf, key), i)
	return e
}

// Caller adds the file:line of the caller with the clog.CallerFieldName key.
// The argument skip is the number of stack frames to ascend
// Skip If not passed, use the global variable CallerSkipFrameCount
func (e *Event) Caller(skip ...int) *Event {
	sk := callerSkipFrameCount
	if len(skip) > 0 {
		sk = skip[0] + callerSkipFrameCount
	}
	return e.caller(sk)
}

func (e *Event) caller(skip int) *Event {
	if e == nil {
		return e
	}
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return e
	}
	e.buf = trs.AppendString(trs.AppendKey(e.buf, callerFieldName), callerMarshalFunc(file, line))
	return e
}

// IPAddr adds IPv4 or IPv6 Address to the event
func (e *Event) IPAddr(key string, ip net.IP) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendIPAddr(trs.AppendKey(e.buf, key), ip)
	return e
}

// IPPrefix adds IPv4 or IPv6 Prefix (address and mask) to the event
func (e *Event) IPPrefix(key string, pfx net.IPNet) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendIPPrefix(trs.AppendKey(e.buf, key), pfx)
	return e
}

// MACAddr adds MAC address to the event
func (e *Event) MACAddr(key string, ha net.HardwareAddr) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendMACAddr(trs.AppendKey(e.buf, key), ha)
	return e
}
