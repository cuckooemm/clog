package clog

func abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}
