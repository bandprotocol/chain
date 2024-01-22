package types

func AbsInt64(x int64) int64 {
	if x < 0 {
		return -1 * x
	}
	return x
}
