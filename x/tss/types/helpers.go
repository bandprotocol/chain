package types

func Uint64ArrayContains(arr []uint64, a uint64) bool {
	for _, v := range arr {
		if v == a {
			return true
		}
	}
	return false
}

func DuplicateInArray(arr []string) bool {
	visited := make(map[string]bool, 0)
	for i := 0; i < len(arr); i++ {
		if visited[arr[i]] {
			return true
		} else {
			visited[arr[i]] = true
		}
	}
	return false
}
