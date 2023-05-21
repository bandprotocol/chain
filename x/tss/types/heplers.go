package types

func Uint64ArrayContains(arr []uint64, a uint64) bool {
	for _, v := range arr {
		if v == a {
			return true
		}
	}

	return false
}
