package client

import (
	"fmt"

	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
)

// MembersResponse wraps the types.Members to provide additional helper methods.
type MembersResponse struct {
	Members []*bandtsstypes.Member
}

// NewMembersResponse creates a new instance of MembersResponse.
func NewMembersResponse(mr *bandtsstypes.QueryMembersResponse) *MembersResponse {
	return &MembersResponse{mr.Members}
}

// FindMembersByAddress finds members in the response by their address.
func (mr MembersResponse) FindMembersByAddress(address string) ([]bandtsstypes.Member, error) {
	members := make([]bandtsstypes.Member, 0)
	for _, member := range mr.Members {
		if member.Address == address {
			members = append(members, *member)
			break
		}
	}

	if len(members) == 0 {
		return members, fmt.Errorf("member not found")
	}

	return members, nil
}
