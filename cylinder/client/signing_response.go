package client

import (
	"fmt"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
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
func (sr *SigningResponse) GetMemberIDs() []tss.MemberID {
	return types.AssignedMembers(sr.Signing.AssignedMembers).MemberIDs()
}

// GetMemberIDs returns all assigned member's id of the assigned members
func (sr *SigningResponse) GetAssignedMember(address string) (types.AssignedMember, error) {
	for _, am := range sr.Signing.AssignedMembers {
		if am.Address == address {
			return am, nil
		}
	}

	return types.AssignedMember{}, fmt.Errorf("%s is not the assigned member", address)
}
