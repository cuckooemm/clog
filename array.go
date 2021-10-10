package clog

import (
	"net"
	"sync"
	"time"
)

var arrayPool = &sync.Pool{
	New: func() interface{} {
		return &Array{
			buf: make([]byte, 0, initCap),
		}
	},
}

// Array 用于预填充数组.
type Array struct {
	buf []byte
}

func putArray(a *Array) {
	if cap(a.buf) > maxCap {
		return
	}
	arrayPool.Put(a)
}

// Arr 创建并添加Array数据到事件上下文.
func Arr() *Array {
	a := arrayPool.Get().(*Array)
	a.buf = a.buf[:0]
	return a
}

// MarshalArray 无操作 - since data is
// already in the needed format.
func (*Array) MarshalArray(*Array) {
}

func (a *Array) write(dst []byte) []byte {
	dst = trs.AppendArrayStart(dst)
	if len(a.buf) > 0 {
		dst = append(dst, a.buf...)
	}
	dst = trs.AppendArrayEnd(dst)
	putArray(a)
	return dst
}

// Object marshals an object that implement the LogObjectMarshaler
// interface and append it to the array.
func (a *Array) Object(obj LogObjectMarshaler) *Array {
	e := Dict()
	obj.MarshalObject(e)
	e.buf = trs.AppendEndMarker(e.buf)
	a.buf = append(trs.AppendArrayDelim(a.buf), e.buf...)
	putEvent(e)
	return a
}

// Str 添加string类型数据到Array.
func (a *Array) Str(val string) *Array {
	a.buf = trs.AppendString(trs.AppendArrayDelim(a.buf), val)
	return a
}

// Bytes 添加[]byte类型数据到Array.
func (a *Array) Bytes(val []byte) *Array {
	a.buf = trs.AppendBytes(trs.AppendArrayDelim(a.buf), val)
	return a
}

// Hex 添加[]byte类型数据，以16进制形式输出到Array
func (a *Array) Hex(val []byte) *Array {
	a.buf = trs.AppendHex(trs.AppendArrayDelim(a.buf), val)
	return a
}

// HexStr 添加string类型数据，以16进制形式输出到Array
func (a *Array) HexStr(val string) *Array {
	a.buf = trs.AppendHex(trs.AppendArrayDelim(a.buf), []byte(val))
	return a
}

// RawJSON 添加原始Json类型数据到Array,同样未对Json格式做检查.
func (a *Array) RawJSON(val []byte) *Array {
	a.buf = append(trs.AppendArrayDelim(a.buf), val...)
	return a
}

// Err 添加序列化后的error数据到Array.
func (a *Array) Err(err error) *Array {
	switch m := errorMarshalFunc(err).(type) {
	case LogObjectMarshaler:
		e := newEvent(nil, 0)
		e.buf = e.buf[:0]
		e.appendObject(m)
		a.buf = append(trs.AppendArrayDelim(a.buf), e.buf...)
		putEvent(e)
	case error:
		if m == nil || isNilValue(m) {
			a.buf = trs.AppendNil(trs.AppendArrayDelim(a.buf))
		} else {
			a.buf = trs.AppendString(trs.AppendArrayDelim(a.buf), m.Error())
		}
	case string:
		a.buf = trs.AppendString(trs.AppendArrayDelim(a.buf), m)
	default:
		a.buf = trs.AppendInterface(trs.AppendArrayDelim(a.buf), m)
	}

	return a
}

// Bool 添加bool类型数据到Array.
func (a *Array) Bool(b bool) *Array {
	a.buf = trs.AppendBool(trs.AppendArrayDelim(a.buf), b)
	return a
}

// Int 添加int类型数据到Array.
func (a *Array) Int(i int) *Array {
	a.buf = trs.AppendInt(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Int8 添加int8类型数据到Array.
func (a *Array) Int8(i int8) *Array {
	a.buf = trs.AppendInt8(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Int16 添加int16类型数据到Array.
func (a *Array) Int16(i int16) *Array {
	a.buf = trs.AppendInt16(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Int32 添加int32类型数据到Array.
func (a *Array) Int32(i int32) *Array {
	a.buf = trs.AppendInt32(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Int64 添加int64类型数据到Array.
func (a *Array) Int64(i int64) *Array {
	a.buf = trs.AppendInt64(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Uint 添加uint类型数据到Array.
func (a *Array) Uint(i uint) *Array {
	a.buf = trs.AppendUint(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Uint8 添加uint8类型数据到Array.
func (a *Array) Uint8(i uint8) *Array {
	a.buf = trs.AppendUint8(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Uint16 添加uint16类型数据到Array.
func (a *Array) Uint16(i uint16) *Array {
	a.buf = trs.AppendUint16(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Uint32 添加uint32类型数据到Array.
func (a *Array) Uint32(i uint32) *Array {
	a.buf = trs.AppendUint32(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Uint64 添加uint64类型数据到Array.
func (a *Array) Uint64(i uint64) *Array {
	a.buf = trs.AppendUint64(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Float32 添加float32类型数据到Array.
func (a *Array) Float32(f float32) *Array {
	a.buf = trs.AppendFloat32(trs.AppendArrayDelim(a.buf), f)
	return a
}

// Float64 添加float64类型数据到Array.
func (a *Array) Float64(f float64) *Array {
	a.buf = trs.AppendFloat64(trs.AppendArrayDelim(a.buf), f)
	return a
}

// Time 添加time类型数据到Array,如需更改默认时间格式通过clog.Set.TimeFormat(timeLayout).
func (a *Array) Time(t time.Time) *Array {
	a.buf = trs.AppendTime(trs.AppendArrayDelim(a.buf), t, timeLayoutFormat)
	return a
}

// Dur 添加time.Duration类型数据到Array.
func (a *Array) Dur(d time.Duration) *Array {
	a.buf = trs.AppendDuration(trs.AppendArrayDelim(a.buf), d)
	return a
}

// Interface 添加任意类型数据到Array.
func (a *Array) Interface(i interface{}) *Array {
	if obj, ok := i.(LogObjectMarshaler); ok {
		return a.Object(obj)
	}
	a.buf = trs.AppendInterface(trs.AppendArrayDelim(a.buf), i)
	return a
}

// IPAddr 添加 IPvo4 r IPv6 类型数据到Array.
func (a *Array) IPAddr(ip net.IP) *Array {
	a.buf = trs.AppendIPAddr(trs.AppendArrayDelim(a.buf), ip)
	return a
}

// IPPrefix 添加 IPv4 or IPv6 (IP + mask) 类型数据到Array.
func (a *Array) IPPrefix(pfx net.IPNet) *Array {
	a.buf = trs.AppendIPPrefix(trs.AppendArrayDelim(a.buf), pfx)
	return a
}

// MACAddr 添加 MAC地址到Array.
func (a *Array) MACAddr(ha net.HardwareAddr) *Array {
	a.buf = trs.AppendMACAddr(trs.AppendArrayDelim(a.buf), ha)
	return a
}
