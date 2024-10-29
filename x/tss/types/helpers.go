package types

import "github.com/bandprotocol/chain/v3/pkg/tss"

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
