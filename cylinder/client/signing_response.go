package client

import (
	"fmt"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// SigningResponse wraps the types.QuerySigningResponse to provide additional helper methods.
type SigningResponse struct {
	types.QuerySigningResponse
}

// NewSigningResponse creates a new instance of SigningResponse.
func NewSigningResponse(gr *types.QuerySigningResponse) *SigningResponse {
	return &SigningResponse{*gr}
}

// GetMemberIDs returns all assigned member's id of the assigned members
// It returns nil if the signing has no current attempt (e.g. it already expired on chain).
func (sr SigningResponse) GetMemberIDs() []tss.MemberID {
	if sr.SigningResult.CurrentSigningAttempt == nil {
		return nil
	}

	assignedMembers := types.AssignedMembers(sr.SigningResult.CurrentSigningAttempt.AssignedMembers)
	return assignedMembers.MemberIDs()
}

// GetAssignedMember returns assigned member of the specific address
func (sr SigningResponse) GetAssignedMember(address string) (types.AssignedMember, error) {
	// The chain drops the current attempt once a signing expires, so a late query has none.
	if sr.SigningResult.CurrentSigningAttempt == nil {
		return types.AssignedMember{}, fmt.Errorf("signing %d has no current attempt", sr.SigningResult.Signing.ID)
	}

	for _, am := range sr.SigningResult.CurrentSigningAttempt.AssignedMembers {
		if am.Address == address {
			return am, nil
		}
	}

	return types.AssignedMember{}, fmt.Errorf("%s is not the assigned member", address)
}
