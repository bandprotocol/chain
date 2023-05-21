package client

import (
	"fmt"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// GroupResponse wraps the types.QueryGroupResponse to provide additional helper methods.
type GroupResponse struct {
	*types.QueryGroupResponse
}

// NewGroupResponse creates a new instance of GroupResponse.
func NewGroupResponse(gr *types.QueryGroupResponse) *GroupResponse {
	return &GroupResponse{gr}
}

// GetRound1Commitment retrieves the Round1Commitment for the specified member ID.
func (gr *GroupResponse) GetRound1Commitment(mid tss.MemberID) (*types.Round1Commitments, error) {
	commitment, ok := gr.AllRound1Commitments[uint64(mid)]
	if !ok {
		return nil, fmt.Errorf("No Round1Commitment from MemberID(%d)", mid)
	}

	return &commitment, nil
}
