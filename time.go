package clog

import (
	"strconv"
	"time"
)

const (
	TimeFormatSec       = ""
	TimeFormatUnixMs    = "UNIXMS"
	TimeFormatUnixMicro = "UNIXMICRO"
)

// AppendTime formats the input time with the given format
// and appends the encoded string to the input byte slice.
func (s transform) AppendTime(dst []byte, t time.Time, format string) []byte {
	switch format {
	case TimeFormatSec:
		return s.AppendInt64(dst, t.Unix())
	case TimeFormatUnixMs:
		return s.AppendInt64(dst, t.UnixNano()/1e6)
	case TimeFormatUnixMicro:
		return s.AppendInt64(dst, t.UnixNano()/1e3)
	}
	return append(t.AppendFormat(append(dst, '"'), format), '"')
}

// AppendDuration formats the input duration with the given unit & format
// and appends the encoded string to the input byte slice.
func (s transform) AppendDuration(dst []byte, d time.Duration) []byte {
	if durationFieldInteger {
		return strconv.AppendInt(dst, int64(d/durationFieldUnit), 10)
	}
	return s.AppendFloat64(dst, float64(d)/float64(durationFieldUnit))
}

// AppendDurations formats the input durations with the given unit & format
// and appends the encoded string list to the input byte slice.
func (s transform) AppendDurations(dst []byte, vals []time.Duration) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = s.AppendDuration(dst, vals[0])
	if len(vals) > 1 {
		for _, d := range vals[1:] {
			dst = s.AppendDuration(append(dst, ','), d)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendUnixTimes(dst []byte, vals []time.Time) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, vals[0].Unix(), 10)
	if len(vals) > 1 {
		for _, t := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), t.Unix(), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendUnixMsTimes(dst []byte, vals []time.Time) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, vals[0].UnixNano()/1e6, 10)
	if len(vals) > 1 {
		for _, t := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), t.UnixNano()/1000000, 10)
		}
	}
	dst = append(dst, ']')
	return dst
}
