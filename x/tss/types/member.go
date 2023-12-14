package types

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
)

// Verify checks if the address of the Member matches the given address
func (m Member) Verify(address string) bool {
	return m.Address == address
}

// Members represents a slice of Member values.
type Members []Member

// GetIDs returns an array of MemberIDs from a collection of members
func (ms Members) GetIDs() []tss.MemberID {
	var mids []tss.MemberID
	for _, member := range ms {
		mids = append(mids, member.ID)
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
