package client

import (
	"fmt"

	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

// MembersResponse wraps the types.Members to provide additional helper methods.
type MembersResponse struct {
	tsstypes.Members
}

// NewMembersResponse creates a new instance of MembersResponse.
func NewMembersResponse(mr *tsstypes.QueryMembersResponse) *MembersResponse {
	return &MembersResponse{mr.Members}
}

// IsActive checks if the member with the given address is active.
func (mr MembersResponse) IsActive(address string) (bool, error) {
	for _, member := range mr.Members {
		if member.Address == address {
			return member.IsActive, nil
		}
	}

	return false, fmt.Errorf("member not found")
}
