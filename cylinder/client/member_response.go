package client

import (
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
)

// MemberResponse is a response from the bandtss client.
type MemberResponse struct {
	Members []bandtsstypes.Member
}

// NewMembersResponse creates a new instance of MembersResponse.
func NewMemberResponse(mr *bandtsstypes.QueryMemberResponse) *MemberResponse {
	members := []bandtsstypes.Member{}
	if mr.CurrentGroupMember.Address != "" {
		members = append(members, mr.CurrentGroupMember)
	}
	if mr.IncomingGroupMember.Address != "" {
		members = append(members, mr.IncomingGroupMember)
	}
	return &MemberResponse{members}
}
