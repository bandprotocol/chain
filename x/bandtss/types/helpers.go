package types

// DuplicateInArray checks if there are any duplicates in the given string array.
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
