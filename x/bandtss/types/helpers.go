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

// ValidMemberStatus returns true if the member status is valid and false otherwise.
func ValidMemberStatus(status MemberStatus) bool {
	if status == MEMBER_STATUS_ACTIVE ||
		status == MEMBER_STATUS_INACTIVE {
		return true
	}
	return false
}
