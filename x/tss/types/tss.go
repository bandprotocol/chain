package types

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
)

// AssignedMembers represents a slice of AssignedMember values.
type AssignedMembers []AssignedMember

// PubDs returns a list of public D points extracted from the AssignedMembers slice.
func (as AssignedMembers) PubDs() (pubDs tss.Points) {
	for _, a := range as {
		pubDs = append(pubDs, a.PubD)
	}
	return
}

// PubEs returns a list of public E points extracted from the AssignedMembers slice.
func (as AssignedMembers) PubEs() (pubEs tss.Points) {
	for _, a := range as {
		pubEs = append(pubEs, a.PubE)
	}
	return
}

// MemberIDs returns a list of MemberIDs extracted from the AssignedMembers slice.
func (as AssignedMembers) MemberIDs() (mids []tss.MemberID) {
	for _, a := range as {
		mids = append(mids, a.MemberID)
	}
	return
}
