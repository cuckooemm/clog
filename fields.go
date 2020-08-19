package clog

import (
	"net"
	"sort"
	"time"
	"unsafe"
)

func isNilValue(i interface{}) bool {
	return (*[2]uintptr)(unsafe.Pointer(&i))[1] == 0
}

func appendFields(dst []byte, fields map[string]interface{}) []byte {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		dst = trs.AppendKey(dst, key)
		val := fields[key]
		if val, ok := val.(LogObjectMarshaler); ok {
			e := newEvent(nil, 0)
			e.buf = e.buf[:0]
			e.appendObject(val)
			dst = append(dst, e.buf...)
			putEvent(e)
			continue
		}
		switch val := val.(type) {
		case string:
			dst = trs.AppendString(dst, val)
		case []byte:
			dst = trs.AppendBytes(dst, val)
		case error:
			switch m := errorMarshalFunc(val).(type) {
			case LogObjectMarshaler:
				e := newEvent(nil, 0)
				e.buf = e.buf[:0]
				e.appendObject(m)
				dst = append(dst, e.buf...)
				putEvent(e)
			case error:
				if m == nil || isNilValue(m) {
					dst = trs.AppendNil(dst)
				} else {
					dst = trs.AppendString(dst, m.Error())
				}
			case string:
				dst = trs.AppendString(dst, m)
			default:
				dst = trs.AppendInterface(dst, m)
			}
		case []error:
			dst = trs.AppendArrayStart(dst)
			for i, err := range val {
				switch m := errorMarshalFunc(err).(type) {
				case LogObjectMarshaler:
					e := newEvent(nil, 0)
					e.buf = e.buf[:0]
					e.appendObject(m)
					dst = append(dst, e.buf...)
					putEvent(e)
				case error:
					if m == nil || isNilValue(m) {
						dst = trs.AppendNil(dst)
					} else {
						dst = trs.AppendString(dst, m.Error())
					}
				case string:
					dst = trs.AppendString(dst, m)
				default:
					dst = trs.AppendInterface(dst, m)
				}

				if i < (len(val) - 1) {
					trs.AppendArrayDelim(dst)
				}
			}
			dst = trs.AppendArrayEnd(dst)
		case bool:
			dst = trs.AppendBool(dst, val)
		case int:
			dst = trs.AppendInt(dst, val)
		case int8:
			dst = trs.AppendInt8(dst, val)
		case int16:
			dst = trs.AppendInt16(dst, val)
		case int32:
			dst = trs.AppendInt32(dst, val)
		case int64:
			dst = trs.AppendInt64(dst, val)
		case uint:
			dst = trs.AppendUint(dst, val)
		case uint8:
			dst = trs.AppendUint8(dst, val)
		case uint16:
			dst = trs.AppendUint16(dst, val)
		case uint32:
			dst = trs.AppendUint32(dst, val)
		case uint64:
			dst = trs.AppendUint64(dst, val)
		case float32:
			dst = trs.AppendFloat32(dst, val)
		case float64:
			dst = trs.AppendFloat64(dst, val)
		case time.Time:
			dst = trs.AppendTime(dst, val, timeLayoutFormat)
		case time.Duration:
			dst = trs.AppendDuration(dst, val, durationFieldUnit, durationFieldInteger)
		case *string:
			if val != nil {
				dst = trs.AppendString(dst, *val)
			} else {
				dst = trs.AppendNil(dst)
			}
		case *bool:
			if val != nil {
				dst = trs.AppendBool(dst, *val)
			} else {
				dst = trs.AppendNil(dst)
			}
		case *int:
			if val != nil {
				dst = trs.AppendInt(dst, *val)
			} else {
				dst = trs.AppendNil(dst)
			}
		case *int8:
			if val != nil {
				dst = trs.AppendInt8(dst, *val)
			} else {
				dst = trs.AppendNil(dst)
			}
		case *int16:
			if val != nil {
				dst = trs.AppendInt16(dst, *val)
			} else {
				dst = trs.AppendNil(dst)
			}
		case *int32:
			if val != nil {
				dst = trs.AppendInt32(dst, *val)
			} else {
				dst = trs.AppendNil(dst)
			}
		case *int64:
			if val != nil {
				dst = trs.AppendInt64(dst, *val)
			} else {
				dst = trs.AppendNil(dst)
			}
		case *uint:
			if val != nil {
				dst = trs.AppendUint(dst, *val)
			} else {
				dst = trs.AppendNil(dst)
			}
		case *uint8:
			if val != nil {
				dst = trs.AppendUint8(dst, *val)
			} else {
				dst = trs.AppendNil(dst)
			}
		case *uint16:
			if val != nil {
				dst = trs.AppendUint16(dst, *val)
			} else {
				dst = trs.AppendNil(dst)
			}
		case *uint32:
			if val != nil {
				dst = trs.AppendUint32(dst, *val)
			} else {
				dst = trs.AppendNil(dst)
			}
		case *uint64:
			if val != nil {
				dst = trs.AppendUint64(dst, *val)
			} else {
				dst = trs.AppendNil(dst)
			}
		case *float32:
			if val != nil {
				dst = trs.AppendFloat32(dst, *val)
			} else {
				dst = trs.AppendNil(dst)
			}
		case *float64:
			if val != nil {
				dst = trs.AppendFloat64(dst, *val)
			} else {
				dst = trs.AppendNil(dst)
			}
		case *time.Time:
			if val != nil {
				dst = trs.AppendTime(dst, *val, timeLayoutFormat)
			} else {
				dst = trs.AppendNil(dst)
			}
		case *time.Duration:
			if val != nil {
				dst = trs.AppendDuration(dst, *val, durationFieldUnit, durationFieldInteger)
			} else {
				dst = trs.AppendNil(dst)
			}
		case []string:
			dst = trs.AppendStrings(dst, val)
		case []bool:
			dst = trs.AppendBools(dst, val)
		case []int:
			dst = trs.AppendInts(dst, val)
		case []int8:
			dst = trs.AppendInts8(dst, val)
		case []int16:
			dst = trs.AppendInts16(dst, val)
		case []int32:
			dst = trs.AppendInts32(dst, val)
		case []int64:
			dst = trs.AppendInts64(dst, val)
		case []uint:
			dst = trs.AppendUints(dst, val)
		//case []uint8:
		//	dst = trs.AppendUints8(dst, val)
		case []uint16:
			dst = trs.AppendUints16(dst, val)
		case []uint32:
			dst = trs.AppendUints32(dst, val)
		case []uint64:
			dst = trs.AppendUints64(dst, val)
		case []float32:
			dst = trs.AppendFloats32(dst, val)
		case []float64:
			dst = trs.AppendFloats64(dst, val)
		case []time.Time:
			dst = trs.AppendTimes(dst, val, timeLayoutFormat)
		case []time.Duration:
			dst = trs.AppendDurations(dst, val, durationFieldUnit, durationFieldInteger)
		case nil:
			dst = trs.AppendNil(dst)
		case net.IP:
			dst = trs.AppendIPAddr(dst, val)
		case net.IPNet:
			dst = trs.AppendIPPrefix(dst, val)
		case net.HardwareAddr:
			dst = trs.AppendMACAddr(dst, val)
		default:
			dst = trs.AppendInterface(dst, val)
		}
	}
	return dst
}
