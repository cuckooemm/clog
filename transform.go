package clog

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strconv"
	"time"
)

type Transform struct{}

var trs = Transform{}

// AppendKey appends a new key to the output.
func (s Transform) AppendKey(dst []byte, key string) []byte {
	if dst[len(dst)-1] != '{' {
		dst = append(dst, ',')
	}
	return append(s.AppendString(dst, key), ':')
}

// AppendHex encodes the input bytes to a hex string and appends
// the encoded string to the input byte slice.
//
// The operation loops though each byte and encodes it as hex using
// the hex lookup table.
func (Transform) AppendHex(dst, s []byte) []byte {
	dst = append(dst, '"')
	for _, v := range s {
		dst = append(dst, hex[v>>4], hex[v&0x0f])
	}
	return append(dst, '"')
}

// AppendBytes is a mirror of appendString with []byte arg
func (Transform) AppendBytes(dst, s []byte) []byte {
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

// AppendBeginMarker inserts a map start into the dst byte array.
func (Transform) AppendBeginMarker(dst []byte) []byte {
	return append(dst, '{')
}

// AppendEndMarker inserts a map end into the dst byte array.
func (Transform) AppendEndMarker(dst []byte) []byte {
	return append(dst, '}')
}

// AppendArrayStart adds markers to indicate the start of an array.
func (Transform) AppendArrayStart(dst []byte) []byte {
	return append(dst, '[')
}

// AppendArrayEnd adds markers to indicate the end of an array.
func (Transform) AppendArrayEnd(dst []byte) []byte {
	return append(dst, ']')
}

// AppendArrayDelim adds markers to indicate end of a particular array element.
func (Transform) AppendArrayDelim(dst []byte) []byte {
	if len(dst) > 0 {
		return append(dst, ',')
	}
	return dst
}

// AppendLineBreak appends a line break.
func (Transform) AppendLineBreak(dst []byte) []byte {
	return append(dst, '\n')
}

// AppendBool converts the input bool to a string and
// appends the encoded string to the input byte slice.
func (Transform) AppendBool(dst []byte, val bool) []byte {
	return strconv.AppendBool(dst, val)
}

// AppendBools encodes the input bools to json and
// appends the encoded string list to the input byte slice.
func (Transform) AppendBools(dst []byte, vals []bool) []byte {
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

// AppendInt converts the input int to a string and
// appends the encoded string to the input byte slice.
func (Transform) AppendInt(dst []byte, val int) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// AppendInts encodes the input ints to json and
// appends the encoded string list to the input byte slice.
func (Transform) AppendInts(dst []byte, vals []int) []byte {
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

// AppendInt8 converts the input []int8 to a string and
// appends the encoded string to the input byte slice.
func (Transform) AppendInt8(dst []byte, val int8) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// AppendInts8 encodes the input int8s to json and
// appends the encoded string list to the input byte slice.
func (Transform) AppendInts8(dst []byte, vals []int8) []byte {
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

// AppendInt16 converts the input int16 to a string and
// appends the encoded string to the input byte slice.
func (Transform) AppendInt16(dst []byte, val int16) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// AppendInts16 encodes the input int16s to json and
// appends the encoded string list to the input byte slice.
func (Transform) AppendInts16(dst []byte, vals []int16) []byte {
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

// AppendInt32 converts the input int32 to a string and
// appends the encoded string to the input byte slice.
func (Transform) AppendInt32(dst []byte, val int32) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// AppendInts32 encodes the input int32s to json and
// appends the encoded string list to the input byte slice.
func (Transform) AppendInts32(dst []byte, vals []int32) []byte {
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

// AppendInt64 converts the input int64 to a string and
// appends the encoded string to the input byte slice.
func (Transform) AppendInt64(dst []byte, val int64) []byte {
	return strconv.AppendInt(dst, val, 10)
}

// AppendInts64 encodes the input int64s to json and
// appends the encoded string list to the input byte slice.
func (Transform) AppendInts64(dst []byte, vals []int64) []byte {
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

// AppendUint converts the input uint to a string and
// appends the encoded string to the input byte slice.
func (Transform) AppendUint(dst []byte, val uint) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// AppendUints encodes the input uints to json and
// appends the encoded string list to the input byte slice.
func (Transform) AppendUints(dst []byte, vals []uint) []byte {
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

// AppendUint8 converts the input uint8 to a string and
// appends the encoded string to the input byte slice.
func (Transform) AppendUint8(dst []byte, val uint8) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// AppendUints8 encodes the input uint8s to json and
// appends the encoded string list to the input byte slice.
func (Transform) AppendUints8(dst []byte, vals []uint8) []byte {
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

// AppendUint16 converts the input uint16 to a string and
// appends the encoded string to the input byte slice.
func (Transform) AppendUint16(dst []byte, val uint16) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// AppendUints16 encodes the input uint16s to json and
// appends the encoded string list to the input byte slice.
func (Transform) AppendUints16(dst []byte, vals []uint16) []byte {
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

// AppendUint32 converts the input uint32 to a string and
// appends the encoded string to the input byte slice.
func (Transform) AppendUint32(dst []byte, val uint32) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// AppendUints32 encodes the input uint32s to json and
// appends the encoded string list to the input byte slice.
func (Transform) AppendUints32(dst []byte, vals []uint32) []byte {
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

// AppendUint64 converts the input uint64 to a string and
// appends the encoded string to the input byte slice.
func (Transform) AppendUint64(dst []byte, val uint64) []byte {
	return strconv.AppendUint(dst, val, 10)
}

// AppendUints64 encodes the input uint64s to json and
// appends the encoded string list to the input byte slice.
func (Transform) AppendUints64(dst []byte, vals []uint64) []byte {
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
	// JSON does not permit NaN or Infinity. A typical JSON encoder would fail
	// with an error, but a logging library wants the data to get thru so we
	// make a tradeoff and store those types as string.
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

// AppendFloat32 converts the input float32 to a string and
// appends the encoded string to the input byte slice.
func (Transform) AppendFloat32(dst []byte, val float32) []byte {
	return appendFloat(dst, float64(val), 32)
}

// AppendFloats32 encodes the input float32s to json and
// appends the encoded string list to the input byte slice.
func (Transform) AppendFloats32(dst []byte, vals []float32) []byte {
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

// AppendFloat64 converts the input float64 to a string and
// appends the encoded string to the input byte slice.
func (Transform) AppendFloat64(dst []byte, val float64) []byte {
	return appendFloat(dst, val, 64)
}

// AppendFloats64 encodes the input float64s to json and
// appends the encoded string list to the input byte slice.
func (Transform) AppendFloats64(dst []byte, vals []float64) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = appendFloat(dst, vals[0], 32)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = appendFloat(append(dst, ','), val, 64)
		}
	}
	dst = append(dst, ']')
	return dst
}

func (s Transform) AppendInterface(dst []byte, i interface{}) []byte {
	marshaled, err := json.Marshal(i)
	if err != nil {
		return s.AppendString(dst, fmt.Sprintf("marshaling error: %v", err))
	}
	return append(dst, marshaled...)
}

// AppendIPAddr adds IPv4 or IPv6 address to dst.
func (s Transform) AppendIPAddr(dst []byte, ip net.IP) []byte {
	return s.AppendString(dst, ip.String())
}

// AppendIPPrefix adds IPv4 or IPv6 Prefix (address & mask) to dst.
func (s Transform) AppendIPPrefix(dst []byte, pfx net.IPNet) []byte {
	return s.AppendString(dst, pfx.String())

}

// AppendMACAddr adds MAC address to dst.
func (s Transform) AppendMACAddr(dst []byte, ha net.HardwareAddr) []byte {
	return s.AppendString(dst, ha.String())
}

// AppendObjectData takes in an object that is already in a byte array
// and adds it to the dst.
func (Transform) AppendObjectData(dst []byte, o []byte) []byte {
	// Three conditions apply here:
	// 1. new content starts with '{' - which should be dropped   OR
	// 2. new content starts with '{' - which should be replaced with ','
	//    to separate with existing content OR
	// 3. existing content has already other fields
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

func (Transform) AppendString(dst []byte, s string) []byte {
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

// AppendNil inserts a 'Nil' object into the dst byte array.
func (Transform) AppendNil(dst []byte) []byte {
	return append(dst, "null"...)
}

func (s Transform) AppendStrings(dst []byte, val []string) []byte {
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

func (s Transform) AppendTimes(dst []byte, val []time.Time, format string) []byte {
	switch format {
	case timeFormatS:
		return appendUnixTimes(dst, val)
	case timeFormatUnixMs:
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
