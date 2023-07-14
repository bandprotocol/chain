package types

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
)

// Verify checks if the address of the Member matches the given address
func (m Member) Verify(address string) bool {
	if m.Address != address {
		return false
	}
	return true
}

// Members represents a slice of Member values.
type Members []Member

// GetIDs returns an array of MemberIDs from a collection of members
func (ms Members) GetIDs() []tss.MemberID {
	var mids []tss.MemberID
	for _, member := range ms {
		mids = append(mids, member.MemberID)
	}

	return mids
}

// HaveMalicious checks if any member in the collection is marked as malicious
func (ms Members) HaveMalicious() bool {
	for _, m := range ms {
		if m.IsMalicious {
			return true
		}
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
