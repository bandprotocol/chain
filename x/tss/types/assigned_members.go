package types

import (
	"bytes"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

// AssignedMembers represents a slice of AssignedMember values.
type AssignedMembers []AssignedMember

// PubDs returns a list of public D points extracted from the AssignedMembers slice.
func (ams AssignedMembers) PubDs() (pubDs tss.Points) {
	for _, am := range ams {
		pubDs = append(pubDs, am.PubD)
	}
	return
}

// PubEs returns a list of public E points extracted from the AssignedMembers slice.
func (ams AssignedMembers) PubEs() (pubEs tss.Points) {
	for _, am := range ams {
		pubEs = append(pubEs, am.PubE)
	}
	return
}

// PubNonces returns a list of public nonce points extracted from the AssignedMembers slice.
func (ams AssignedMembers) PubNonces() (pubNonces tss.Points) {
	for _, am := range ams {
		pubNonces = append(pubNonces, am.PubNonce)
	}
	return
}

// MemberIDs returns a list of MemberIDs extracted from the AssignedMembers slice.
func (ams AssignedMembers) MemberIDs() (mids []tss.MemberID) {
	for _, am := range ams {
		mids = append(mids, am.MemberID)
	}
	return
}

// FindAssignedMember searches for an assigned member in the AssignedMembers slice with a matching MemberID.
func (ams AssignedMembers) FindAssignedMember(mid tss.MemberID) (AssignedMember, bool) {
	for _, am := range ams {
		if mid == am.MemberID {
			return am, true
		}
	}
	return AssignedMember{}, false
}

// VerifySignatureR checks if a given MemberID has a matching public nonce (R) in the AssignedMembers slice.
func (ams AssignedMembers) VerifySignatureR(mid tss.MemberID, r tss.Point) bool {
	for _, am := range ams {
		if mid == am.MemberID {
			return bytes.Equal(r, am.PubNonce)
		}
	}
	return false
}
