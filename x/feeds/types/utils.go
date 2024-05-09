package types

// AbsInt64 returns an absolute of int64.
// Panics on min int64 (-9223372036854775808).
func AbsInt64(x int64) int64 {
	if x < 0 {
		return -1 * x
	}
	return x
}
