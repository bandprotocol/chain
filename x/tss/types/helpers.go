package types

// Uint64ArrayContains checks if the given array contains the specified uint64 value.
func Uint64ArrayContains(arr []uint64, a uint64) bool {
	for _, v := range arr {
		if v == a {
			return true
		}
	}
	return false
}

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

// ValidMemberStatus returns true if the member status is valid and false
// otherwise.
func ValidMemberStatus(status MemberStatus) bool {
	if status == MEMBER_STATUS_ACTIVE ||
		status == MEMBER_STATUS_INACTIVE ||
		status == MEMBER_STATUS_JAIL {
		return true
	}
	return false
}
