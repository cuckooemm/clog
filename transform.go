package clog

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strconv"
	"time"
)

type transform struct{}

var trs = transform{}

// AppendKey 添加key到bytes.
func (s transform) AppendKey(dst []byte, key string) []byte {
	if dst[len(dst)-1] != '{' {
		dst = append(dst, ',')
	}
	return append(s.AppendString(dst, key), ':')
}

// AppendHex 以16进制形式添加[]byte到bytes.
func (transform) AppendHex(dst, s []byte) []byte {
	dst = append(dst, '"')
	for _, v := range s {
		dst = append(dst, hex[v>>4], hex[v&0x0f])
	}
	return append(dst, '"')
}

// AppendBytes 等同于 appendString
func (transform) AppendBytes(dst, s []byte) []byte {
	dst = append(dst, '"')
	for i := 0; i < len(s); i++ {
		if !noEscapeTable[s[i]] {
			dst = appendBytesComplex(dst, s, i)
			return append(dst, '"')
		}
	}
	dst = append(dst, s...)
	return append(dst, '"')
}

// AppendBeginMarker 添加Json开始标记.
func (transform) AppendBeginMarker(dst []byte) []byte {
	return append(dst, '{')
}

// AppendEndMarker 添加Json结束标记.
func (transform) AppendEndMarker(dst []byte) []byte {
	return append(dst, '}')
}

// AppendArrayStart 添加数组开始标记.
func (transform) AppendArrayStart(dst []byte) []byte {
	return append(dst, '[')
}

// AppendArrayEnd 添加数组结束标记.
func (transform) AppendArrayEnd(dst []byte) []byte {
	return append(dst, ']')
}

// AppendArrayDelim 添加数组数据分隔符.
func (transform) AppendArrayDelim(dst []byte) []byte {
	if len(dst) > 0 {
		return append(dst, ',')
	}
	return dst
}

// AppendLineBreak 添加换行符.
func (transform) AppendLineBreak(dst []byte) []byte {
	return append(dst, '\n')
}

// AppendBool 转换bool为string至bytes.
func (transform) AppendBool(dst []byte, val bool) []byte {
	return strconv.AppendBool(dst, val)
}

// AppendBools 添加[]bool为[]string至bytes.
func (transform) AppendBools(dst []byte, vals []bool) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendBool(dst, vals[0])
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendBool(append(dst, ','), val)
		}
	}
	dst = append(dst, ']')
	return dst
}

// AppendInt 添加int类型到bytes.
func (transform) AppendInt(dst []byte, val int) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// AppendInts 添加[]int类型到bytes.
func (transform) AppendInts(dst []byte, vals []int) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, int64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), int64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

// AppendInt8 添加int8类型到bytes.
func (transform) AppendInt8(dst []byte, val int8) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// AppendInts8 添加[]int8类型到bytes.
func (transform) AppendInts8(dst []byte, vals []int8) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, int64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), int64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

// AppendInt16 添加int16类型到bytes.
func (transform) AppendInt16(dst []byte, val int16) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// AppendInts16 添加[]int16类型到bytes.
func (transform) AppendInts16(dst []byte, vals []int16) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, int64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), int64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

// AppendInt32 添加int32类型到bytes.
func (transform) AppendInt32(dst []byte, val int32) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// AppendInts32 添加[]int32类型到bytes.
func (transform) AppendInts32(dst []byte, vals []int32) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, int64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), int64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

// AppendInt64 添加int64类型到bytes.
func (transform) AppendInt64(dst []byte, val int64) []byte {
	return strconv.AppendInt(dst, val, 10)
}

// AppendInts64 添加[]int64类型到bytes.
func (transform) AppendInts64(dst []byte, vals []int64) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, vals[0], 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), val, 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

// AppendUint 添加uint类型到bytes.
func (transform) AppendUint(dst []byte, val uint) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// AppendUints 添加uint类型到bytes.
func (transform) AppendUints(dst []byte, vals []uint) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendUint(dst, uint64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendUint(append(dst, ','), uint64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

// AppendUint8 添加uint8类型到bytes.
func (transform) AppendUint8(dst []byte, val uint8) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// AppendUints8 添加[]uint8类型到bytes.
func (transform) AppendUints8(dst []byte, vals []uint8) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendUint(dst, uint64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendUint(append(dst, ','), uint64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

// AppendUint16 添加uint16类型到bytes.
func (transform) AppendUint16(dst []byte, val uint16) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// AppendUints16 添加[]uint16类型到bytes.
func (transform) AppendUints16(dst []byte, vals []uint16) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendUint(dst, uint64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendUint(append(dst, ','), uint64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

// AppendUint32 添加uint32类型到bytes.
func (transform) AppendUint32(dst []byte, val uint32) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// AppendUints32 添加[]uint32类型到bytes.
func (transform) AppendUints32(dst []byte, vals []uint32) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendUint(dst, uint64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendUint(append(dst, ','), uint64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

// AppendUint64 添加uint64类型到bytes.
func (transform) AppendUint64(dst []byte, val uint64) []byte {
	return strconv.AppendUint(dst, val, 10)
}

// AppendUints64 添加[]uint64类型到bytes.
func (transform) AppendUints64(dst []byte, vals []uint64) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendUint(dst, vals[0], 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendUint(append(dst, ','), val, 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendFloat(dst []byte, val float64, bitSize int) []byte {
	switch {
	case math.IsNaN(val):
		return append(dst, `"NaN"`...)
	case math.IsInf(val, 1):
		return append(dst, `"+Inf"`...)
	case math.IsInf(val, -1):
		return append(dst, `"-Inf"`...)
	}
	return strconv.AppendFloat(dst, val, 'f', -1, bitSize)
}

// AppendFloat32 添加float32类型到bytes.
func (transform) AppendFloat32(dst []byte, val float32) []byte {
	return appendFloat(dst, float64(val), 32)
}

// AppendFloats32 添加[]float32类型到bytes.
func (transform) AppendFloats32(dst []byte, vals []float32) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = appendFloat(dst, float64(vals[0]), 32)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = appendFloat(append(dst, ','), float64(val), 32)
		}
	}
	dst = append(dst, ']')
	return dst
}

// AppendFloat64 添加float64类型到bytes.
func (transform) AppendFloat64(dst []byte, val float64) []byte {
	return appendFloat(dst, val, 64)
}

// AppendFloats64 添加[]float64类型到bytes.
func (transform) AppendFloats64(dst []byte, vals []float64) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = appendFloat(dst, vals[0], 64)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = appendFloat(append(dst, ','), val, 64)
		}
	}
	dst = append(dst, ']')
	return dst
}

// AppendInterface 添加interface{}类型到bytes.
func (s transform) AppendInterface(dst []byte, i interface{}) []byte {
	marshaled, err := json.Marshal(i)
	if err != nil {
		return s.AppendString(dst, fmt.Sprintf("marshaling error: %v", err))
	}
	return append(dst, marshaled...)
}

// AppendIPAddr 添加 IPv4 or IPv6地址到Slice.
func (s transform) AppendIPAddr(dst []byte, ip net.IP) []byte {
	return s.AppendString(dst, ip.String())
}

// AppendIPPrefix 添加 IPv4 or IPv6 Prefix (address & mask)到slice.
func (s transform) AppendIPPrefix(dst []byte, pfx net.IPNet) []byte {
	return s.AppendString(dst, pfx.String())

}

// AppendMACAddr 添加 MAC 硬件地址到Slice.
func (s transform) AppendMACAddr(dst []byte, ha net.HardwareAddr) []byte {
	return s.AppendString(dst, ha.String())
}

func (transform) AppendObjectData(dst []byte, o []byte) []byte {
	if o[0] == '{' {
		if len(dst) > 1 {
			dst = append(dst, ',')
		}
		o = o[1:]
	} else if len(dst) > 1 {
		dst = append(dst, ',')
	}
	return append(dst, o...)
}

func (transform) AppendString(dst []byte, s string) []byte {
	// Start with a double quote.
	dst = append(dst, '"')
	// Loop through each character in the string.
	for i := 0; i < len(s); i++ {
		// Check if the character needs encoding. Control characters, slashes,
		// and the double quote need json encoding. Bytes above the ascii
		// boundary needs utf8 encoding.
		if !noEscapeTable[s[i]] {
			// We encountered a character that needs to be encoded. Switch
			// to complex version of the algorithm.
			dst = appendStringComplex(dst, s, i)
			return append(dst, '"')
		}
	}
	// The string has no need for encoding an therefore is directly
	// appended to the byte slice.
	dst = append(dst, s...)
	// End with a double quote
	return append(dst, '"')
}

// AppendNil 添加`nil`类型到bytes.
func (transform) AppendNil(dst []byte) []byte {
	return append(dst, "null"...)
}

func (s transform) AppendStrings(dst []byte, val []string) []byte {
	if len(val) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = s.AppendString(dst, val[0])
	if len(val) > 1 {
		for _, val := range val[1:] {
			dst = s.AppendString(append(dst, ','), val)
		}
	}
	dst = append(dst, ']')
	return dst
}

func (s transform) AppendTimes(dst []byte, val []time.Time, format string) []byte {
	switch format {
	case TimeFormatUnixSec:
		return appendUnixTimes(dst, val)
	case TimeFormatUnixMs:
		return appendUnixMsTimes(dst, val)
	}
	if len(val) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = append(val[0].AppendFormat(append(dst, '"'), format), '"')
	if len(val) > 1 {
		for _, t := range val[1:] {
			dst = append(t.AppendFormat(append(dst, ',', '"'), format), '"')
		}
	}
	dst = append(dst, ']')
	return dst
}
