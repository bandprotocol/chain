package types

import "github.com/bandprotocol/chain/v2/pkg/tss"

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

// ValidReplacementStatus returns true if the replacement group status is valid and false
// otherwise.
func ValidReplacementStatus(status ReplacementStatus) bool {
	if status == REPLACEMENT_STATUS_WAITING ||
		status == REPLACEMENT_STATUS_FALLEN ||
		status == REPLACEMENT_STATUS_SUCCESS {
		return true
	}
	return false
}

// FindMemberSlot is used to figure out the position of 'to' within an array.
// This array follows a pattern defined by a rule (f_i(j)), where j ('to') != i ('from').
func FindMemberSlot(from tss.MemberID, to tss.MemberID) tss.MemberID {
	// Convert 'to' to 0-indexed system
	slot := to - 1

	// If 'from' is less than 'to', subtract 1 again
	if from < to {
		slot--
	}

	return slot
}
