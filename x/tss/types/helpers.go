package types

import "github.com/bandprotocol/chain/v2/pkg/tss"

// VerifyMember checks if the given member's address matches the provided address.
func VerifyMember(member Member, address string) bool {
	if member.Address != address {
		return false
	}
	return true
}

// FindMemberSlot calculates the slot value based on 'from' and 'to' MemberIDs.
func FindMemberSlot(from tss.MemberID, to tss.MemberID) tss.MemberID {
	slot := to - 1
	if from < to {
		slot--
	}

	return slot
}

// GetMemberIDs get the list of the member ID from all members.
func GetMemberIDs(members []Member) []tss.MemberID {
	var mids []tss.MemberID
	for _, member := range members {
		mids = append(mids, member.MemberID)
	}
	return mids
}

// HaveMalicious checks if any member in the given slice is marked as malicious.
func HaveMalicious(members []Member) bool {
	for _, m := range members {
		if m.IsMalicious {
			return true
		}
	}

	return false
}

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
