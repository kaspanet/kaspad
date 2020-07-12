package math

// MinInt returns the smaller of x or y.
func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// MinUint32 returns the smaller of x or y.
func MinUint32(x, y uint32) uint32 {
	if x < y {
		return x
	}
	return y
}
