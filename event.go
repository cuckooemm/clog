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
	stack bool // 错误堆栈跟踪
	hook  []Hook
}

type LogObjectMarshaler interface {
	MarshalObject(e *Event)
}

type LogArrayMarshaler interface {
	MarshalArray(a *Array)
}

func newEvent(w LevelWriter, level Level) *Event {
	e := eventPool.Get().(*Event)
	e.buf = e.buf[:0]
	e.hook = nil
	e.buf = trs.AppendBeginMarker(e.buf)
	e.w = w
	e.level = level
	e.stack = false
	return e
}

func putEvent(e *Event) {
	if cap(e.buf) > maxCap {
		return
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

// IsEnabled 判断此次事件是否已被关闭 返回 true 则表示正常输出  false 则为已关闭输出
// 返回true 并不代表一定输出， 还需看日志等级
func (e *Event) IsEnabled() bool {
	return e != nil && e.level != Disabled
}

// Discard 关闭此条日志输出
func (e *Event) Discard() *Event {
	if e == nil {
		return e
	}
	e.level = Disabled
	return nil
}

// Msg 输出日志 如果参数字段不为空字符串  则增加 messageFieldName 定义的field name
// 调用后输出此次日志上下文数据
// NOTICE: 此方法只能被调用一次，多次调用会引发意料之外的结果
func (e *Event) Msg(msg string) {
	if e == nil {
		return
	}
	e.msg(msg)
}

// Cease 等同于调用Msg("")
func (e *Event) Cease() {
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
	for _, hook := range e.hook {
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

// Fields 向事件上下文添加Map类型数据，
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

// EmbedObject 序列化实现了 LogObjectMarshaler interface 的类型数据到事件上下文.
func (e *Event) EmbedObject(obj LogObjectMarshaler) *Event {
	if e == nil {
		return e
	}
	obj.MarshalObject(e)
	return e
}

// Str 添加String类型数据到事件上下文.
func (e *Event) Str(key, val string) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendString(trs.AppendKey(e.buf, key), val)
	return e
}

// Strs 添加[]String类型数据到事件上下文.
func (e *Event) Strs(key string, vals []string) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendStrings(trs.AppendKey(e.buf, key), vals)
	return e
}

// Stringer 添加实现`String()`接口的类型数据到事件上下文. 如果为nil 则输出 null.
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

// Bytes 添加[]byte类型数据到事件上下文. 如果为nil 则输出空字符.
//	Log.Bytes("bytes", []byte("bar")).Cease()
// Output:
//	{"bytes":"bar"}
// NOTE:
//	非ASCII字符将被16进制编码输出
func (e *Event) Bytes(key string, val []byte) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendBytes(trs.AppendKey(e.buf, key), val)
	return e
}

// Hex 添加以16进制编码的[]byte类型数据到事件上下文.
func (e *Event) Hex(key string, val []byte) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendHex(trs.AppendKey(e.buf, key), val)
	return e
}

// HexStr 同Hex().
func (e *Event) HexStr(key string, val string) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendHex(trs.AppendKey(e.buf, key), []byte(val))
	return e
}

// RawJSON 添加原始JSON数据到事件上下文.
//	Log().RawJSON(`{"some":"json"}`).Cease()
// NOTE:
// 		未对数据做Json格式校验
// Output:
//	{"some":"json"}
func (e *Event) RawJSON(key string, b []byte) *Event {
	if e == nil {
		return e
	}
	e.buf = append(trs.AppendKey(e.buf, key), b...)
	return e
}

// AnErr 添加序列化后的error到事件上下文.
// 如果err为nil，则field不会被添加.
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

// Errs 添加Err数组到事件上下文.
// 通过clog.Set.ErrMarshalHandler(func)自定义err输出.
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

// Err 向时间上下文添加error信息，如果err为nil，则不添加field。
//
// 通过clog.Set.FiledName().ErrorFieldName("")更改默认 field name.
//
// 如果在此之前调用了Stack()函数且通过clog.Set.ErrStackMarshal(func)定义了errorStackMarshal
// 则错误通过errorStackMarshal将结果添加到事件上下文中.
// 通过clog.Set.FiledName().ErrStackFieldName()更改默认 field name.
func (e *Event) Err(err error) *Event {
	if e == nil {
		return e
	}
	if e.stack && errorStackMarshal != nil {
		switch m := errorStackMarshal(err).(type) {
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

// Stack 为传递给Err()的错误开启堆栈打印跟踪.
//
// 通过设置clog.Set.ErrStackMarshal(func())使此方法打印某些操作.
func (e *Event) Stack() *Event {
	if e == nil {
		return e
	}
	e.stack = true
	return e
}

// Bool 添加bool类型数据到事件上下文.
// 		Log().Bool("true",true).Bool("false",false).Cease()
// Output:
//  	{"true":true,"false":false}
func (e *Event) Bool(key string, b bool) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendBool(trs.AppendKey(e.buf, key), b)
	return e
}

// Bools 添加[]bool类型数据到事件上下文.
//  	Log().Bools("bool",[]bool{true,false}).Cease()
// Output:
//  	{"bool":[true,false]}
func (e *Event) Bools(key string, b []bool) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendBools(trs.AppendKey(e.buf, key), b)
	return e
}

// Int 添加int类型数据到事件上下文.
//  	Log().Int("int",1).Cease()
// Output:
//  	{"int":1}
func (e *Event) Int(key string, i int) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInt(trs.AppendKey(e.buf, key), i)
	return e
}

// Ints 添加[]int类型数据到事件上下文.
//  	Log().Int("int",[]int{1,2,3}).Cease()
// Output:
//  	{"int":[1,2,3]]}
func (e *Event) Ints(key string, i []int) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInts(trs.AppendKey(e.buf, key), i)
	return e
}

// Int8 添加int8类型数据到事件上下文.
func (e *Event) Int8(key string, i int8) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInt8(trs.AppendKey(e.buf, key), i)
	return e
}

// Ints8 添加[]int8类型数据到事件上下文.
func (e *Event) Ints8(key string, i []int8) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInts8(trs.AppendKey(e.buf, key), i)
	return e
}

// Int16 添加int16类型数据到事件上下文.
func (e *Event) Int16(key string, i int16) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInt16(trs.AppendKey(e.buf, key), i)
	return e
}

// Ints16 添加[]int16类型数据到事件上下文.
func (e *Event) Ints16(key string, i []int16) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInts16(trs.AppendKey(e.buf, key), i)
	return e
}

// Int32 添加int32类型数据到事件上下文.
func (e *Event) Int32(key string, i int32) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInt32(trs.AppendKey(e.buf, key), i)
	return e
}

// Ints32 添加[]int32类型数据到事件上下文.
func (e *Event) Ints32(key string, i []int32) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInts32(trs.AppendKey(e.buf, key), i)
	return e
}

// Int64 添加int64类型数据到事件上下文.
func (e *Event) Int64(key string, i int64) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInt64(trs.AppendKey(e.buf, key), i)
	return e
}

// Ints64 添加[]int64类型数据到事件上下文.
func (e *Event) Ints64(key string, i []int64) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendInts64(trs.AppendKey(e.buf, key), i)
	return e
}

// Uint 添加uint类型数据到事件上下文.
func (e *Event) Uint(key string, i uint) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUint(trs.AppendKey(e.buf, key), i)
	return e
}

// Uints 添加[]uint类型数据到事件上下文.
func (e *Event) Uints(key string, i []uint) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUints(trs.AppendKey(e.buf, key), i)
	return e
}

// Uint8 添加uint8类型数据到事件上下文.
func (e *Event) Uint8(key string, i uint8) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUint8(trs.AppendKey(e.buf, key), i)
	return e
}

// Uints8 添加[]uint8类型数据到事件上下文.
func (e *Event) Uints8(key string, i []uint8) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUints8(trs.AppendKey(e.buf, key), i)
	return e
}

// Uint16 添加uint16类型数据到事件上下文.
func (e *Event) Uint16(key string, i uint16) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUint16(trs.AppendKey(e.buf, key), i)
	return e
}

// Uints16 添加[]uint16类型数据到事件上下文.
func (e *Event) Uints16(key string, i []uint16) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUints16(trs.AppendKey(e.buf, key), i)
	return e
}

// Uint32 添加uint32类型数据到事件上下文.
func (e *Event) Uint32(key string, i uint32) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUint32(trs.AppendKey(e.buf, key), i)
	return e
}

// Uints32 添加[]uint32类型数据到事件上下文.
func (e *Event) Uints32(key string, i []uint32) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUints32(trs.AppendKey(e.buf, key), i)
	return e
}

// Uint64 添加uint64类型数据到事件上下文.
func (e *Event) Uint64(key string, i uint64) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUint64(trs.AppendKey(e.buf, key), i)
	return e
}

// Uints64 添加[]uint64类型数据到事件上下文.
func (e *Event) Uints64(key string, i []uint64) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendUints64(trs.AppendKey(e.buf, key), i)
	return e
}

// Float32 添加Float32类型数据到事件上下文.
func (e *Event) Float32(key string, f float32) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendFloat32(trs.AppendKey(e.buf, key), f)
	return e
}

// Floats32 添加[]Float32类型数据到事件上下文.
func (e *Event) Floats32(key string, f []float32) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendFloats32(trs.AppendKey(e.buf, key), f)
	return e
}

// Float64 添加Float64类型数据到事件上下文.
func (e *Event) Float64(key string, f float64) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendFloat64(trs.AppendKey(e.buf, key), f)
	return e
}

// Floats64 添加Float64类型数据到事件上下文.
func (e *Event) Floats64(key string, f []float64) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendFloats64(trs.AppendKey(e.buf, key), f)
	return e
}

// Timestamp 添加时间数据到事件上下文.
// 通过clog.Set.FiledName().TimestampFieldName()更改timestampFieldName. 函数等同于调用Time("time",time.Now())
// 通过clog.Set.TimeFormat(timeLayout)更改timeLayoutFormat以自定义默认时间格式
//
//  	Log().Timestamp().Cease()
//  	Log().Timestamp().Timestamp().Cease()
// NOTE:
//	函数调用多次并不会对结果去重
// Output:
//      {"time":"2006-01-02T15:04:05+08:00"}
//      {"time":"2006-01-02T15:04:05+08:00","time":"2006-01-02T15:04:05+08:00"}
func (e *Event) Timestamp() *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendTime(trs.AppendKey(e.buf, timestampFieldName), timestampFunc(), timeLayoutFormat)
	return e
}

// Time 添加time类型数据到事件上下文.
// 更改timeLayoutFormat以自定义默认时间格式
//  	Log().Time("t",time.Now()).Cease()
// Output:
//  	{"t":"2006-01-02T15:04:05+08:00"}
func (e *Event) Time(key string, t time.Time) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendTime(trs.AppendKey(e.buf, key), t, timeLayoutFormat)
	return e
}

// TimeF 添加time类型自定义时间格式数据到事件上下文.
//  	Log().TimeF("x",time.Now(),time.RFC822Z).Cease()
// Output:
//  	{"x":"02 Jan 06 15:04 -0700"}
func (e *Event) TimeF(key string, t time.Time, format string) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendTime(trs.AppendKey(e.buf, key), t, format)
	return e
}

// Times 添加[]time类型数据到事件上下文，同Time().
func (e *Event) Times(key string, t []time.Time) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendTimes(trs.AppendKey(e.buf, key), t, timeLayoutFormat)
	return e
}

// TimeDur 添加time.Duration类型数据事件上下文. 默认以毫秒(ms)为单位，输出float类型数据
// clog.Set.BaseTimeDurationInteger() 设置以整数类型输出
//	Log().TimeDur("dur",time.Second + 300 * time.Millisecond + 300 * time.Microsecond).Cease()
//	clog.Set.BaseTimeDurationInteger().BaseTimeDurationUnit(time.Microsecond)
//	Log().TimeDur("dur",time.Second + 300 * time.Millisecond + 300 * time.Microsecond).Cease()
// Output:
// 	{"dur":1300.3}
//	{"dur":1300300}
// NOTE:
//	Set.BaseTimeDurationInteger()与Set.BaseTimeDurationUnit()函数非线程安全，仅在初始化时使用
func (e *Event) TimeDur(key string, d time.Duration) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendDuration(trs.AppendKey(e.buf, key), d)
	return e
}

// TimeDurs 添加[]time.Duration类型数据到事件上下文，同TimeDur.
func (e *Event) TimeDurs(key string, d []time.Duration) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendDurations(trs.AppendKey(e.buf, key), d)
	return e
}

// TimeDurStr 添加[]Time.Duration类型数据到事件上下文,输出String类型.
//	Log().TimeDurStr("dur",time.Second + 300 * time.Millisecond + 300 * time.Microsecond).Cease()
// Output:
//	{"dur":"1.3003s"}
func (e *Event) TimeDurStr(key string, d time.Duration) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendString(trs.AppendKey(e.buf, key), d.String())
	return e
}

// TimeDiff 添加start(param1)和end(param2)时间差值到事件上下文,输出格式与单位同TimeDur().
//
// 如果 start < end 则输出正数，反之为负数
//	t1 := time.Date(2021, 9, 28, 12, 23, 22, 0, time.Local)
//	t2 := time.Date(2021, 9, 28, 12, 24, 44, 0, time.Local)
//	Log().TimeDiff("dur",t1,t2).Cease()
// Output:
//	{"t":82000}
func (e *Event) TimeDiff(key string, start time.Time, end time.Time) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendDuration(trs.AppendKey(e.buf, key), end.Sub(start))
	return e
}

// TimeDiffStr 添加start(param1)和end(param2)时间差值到事件上下文,同TimeDiff(); 输出格式与单位同TimeDurStr().
func (e *Event) TimeDiffStr(key string, start time.Time, end time.Time) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendString(trs.AppendKey(e.buf, key), end.Sub(start).String())
	return e
}

// Interface 添加interface{}类型数据到事件上下文.
//
// 如果实现了 LogObjectMarshaler{}接口的MarshalObject(e *Event)方法，则使用序列化方法，否则使用jsonEncode.
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

// Caller 添加函数调用文件与行号信息到事件上下文.
// 通过 clog.Set.FiledName().CallerFieldName("")更改默认field name.
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

// IPAddr 添加 IPv4 or IPv6 地址到事件上下文.
func (e *Event) IPAddr(key string, ip net.IP) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendIPAddr(trs.AppendKey(e.buf, key), ip)
	return e
}

// IPPrefix 添加 IPv4 or IPv6 Prefix (address and mask) 到事件上下文.
func (e *Event) IPPrefix(key string, pfx net.IPNet) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendIPPrefix(trs.AppendKey(e.buf, key), pfx)
	return e
}

// MACAddr 添加 MAC 地址到事件上下文.
func (e *Event) MACAddr(key string, ha net.HardwareAddr) *Event {
	if e == nil {
		return e
	}
	e.buf = trs.AppendMACAddr(trs.AppendKey(e.buf, key), ha)
	return e
}
