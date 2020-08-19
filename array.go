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

// Array is used to prepopulate an array of items
// which can be re-used to add to log messages.
type Array struct {
	buf []byte
}

func putArray(a *Array) {
	if cap(a.buf) > maxCap {
		//return
		a.buf = a.buf[:0:initCap]
	}
	arrayPool.Put(a)
}

// Arr creates an array to be added to an Event or Context.
func Arr() *Array {
	a := arrayPool.Get().(*Array)
	a.buf = a.buf[:0]
	return a
}

// MarshalClogArray method here is no-op - since data is
// already in the needed format.
func (*Array) MarshalArray(*Array) {
}

func (a *Array) write(dst []byte) []byte {
	dst = trs.AppendArrayStart(dst)
	if len(a.buf) > 0 {
		dst = append(append(dst, a.buf...))
	}
	dst = trs.AppendArrayEnd(dst)
	putArray(a)
	return dst
}

// Object marshals an object that implement the LogObjectMarshaler
// interface and append append it to the array.
func (a *Array) Object(obj LogObjectMarshaler) *Array {
	e := Dict()
	obj.MarshalObject(e)
	e.buf = trs.AppendEndMarker(e.buf)
	a.buf = append(trs.AppendArrayDelim(a.buf), e.buf...)
	putEvent(e)
	return a
}

// Str append append the val as a string to the array.
func (a *Array) Str(val string) *Array {
	a.buf = trs.AppendString(trs.AppendArrayDelim(a.buf), val)
	return a
}

// Bytes append append the val as a string to the array.
func (a *Array) Bytes(val []byte) *Array {
	a.buf = trs.AppendBytes(trs.AppendArrayDelim(a.buf), val)
	return a
}

// Hex append append the val as a hex string to the array.
func (a *Array) Hex(val []byte) *Array {
	a.buf = trs.AppendHex(trs.AppendArrayDelim(a.buf), val)
	return a
}

// RawJSON adds already trsoded JSON to the array.
func (a *Array) RawJSON(val []byte) *Array {
	a.buf = append(trs.AppendArrayDelim(a.buf), val...)
	return a
}

// Err serializes and appends the err to the array.
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

// Bool append append the val as a bool to the array.
func (a *Array) Bool(b bool) *Array {
	a.buf = trs.AppendBool(trs.AppendArrayDelim(a.buf), b)
	return a
}

// Int append append i as a int to the array.
func (a *Array) Int(i int) *Array {
	a.buf = trs.AppendInt(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Int8 append append i as a int8 to the array.
func (a *Array) Int8(i int8) *Array {
	a.buf = trs.AppendInt8(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Int16 append append i as a int16 to the array.
func (a *Array) Int16(i int16) *Array {
	a.buf = trs.AppendInt16(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Int32 append append i as a int32 to the array.
func (a *Array) Int32(i int32) *Array {
	a.buf = trs.AppendInt32(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Int64 append append i as a int64 to the array.
func (a *Array) Int64(i int64) *Array {
	a.buf = trs.AppendInt64(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Uint append append i as a uint to the array.
func (a *Array) Uint(i uint) *Array {
	a.buf = trs.AppendUint(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Uint8 append append i as a uint8 to the array.
func (a *Array) Uint8(i uint8) *Array {
	a.buf = trs.AppendUint8(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Uint16 append append i as a uint16 to the array.
func (a *Array) Uint16(i uint16) *Array {
	a.buf = trs.AppendUint16(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Uint32 append append i as a uint32 to the array.
func (a *Array) Uint32(i uint32) *Array {
	a.buf = trs.AppendUint32(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Uint64 append append i as a uint64 to the array.
func (a *Array) Uint64(i uint64) *Array {
	a.buf = trs.AppendUint64(trs.AppendArrayDelim(a.buf), i)
	return a
}

// Float32 append append f as a float32 to the array.
func (a *Array) Float32(f float32) *Array {
	a.buf = trs.AppendFloat32(trs.AppendArrayDelim(a.buf), f)
	return a
}

// Float64 append append f as a float64 to the array.
func (a *Array) Float64(f float64) *Array {
	a.buf = trs.AppendFloat64(trs.AppendArrayDelim(a.buf), f)
	return a
}

// Time append append t formated as string using clog.timeFieldFormat.
func (a *Array) Time(t time.Time) *Array {
	a.buf = trs.AppendTime(trs.AppendArrayDelim(a.buf), t, timeLayoutFormat)
	return a
}

// Dur append append d to the array.
func (a *Array) Dur(d time.Duration) *Array {
	a.buf = trs.AppendDuration(trs.AppendArrayDelim(a.buf), d, durationFieldUnit, durationFieldInteger)
	return a
}

// Interface append append i marshaled using reflection.
func (a *Array) Interface(i interface{}) *Array {
	if obj, ok := i.(LogObjectMarshaler); ok {
		return a.Object(obj)
	}
	a.buf = trs.AppendInterface(trs.AppendArrayDelim(a.buf), i)
	return a
}

// IPAddr adds IPv4 or IPv6 address to the array
func (a *Array) IPAddr(ip net.IP) *Array {
	a.buf = trs.AppendIPAddr(trs.AppendArrayDelim(a.buf), ip)
	return a
}

// IPPrefix adds IPv4 or IPv6 Prefix (IP + mask) to the array
func (a *Array) IPPrefix(pfx net.IPNet) *Array {
	a.buf = trs.AppendIPPrefix(trs.AppendArrayDelim(a.buf), pfx)
	return a
}

// MACAddr adds a MAC (Ethernet) address to the array
func (a *Array) MACAddr(ha net.HardwareAddr) *Array {
	a.buf = trs.AppendMACAddr(trs.AppendArrayDelim(a.buf), ha)
	return a
}
