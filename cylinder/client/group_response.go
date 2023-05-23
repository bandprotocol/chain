package client

import (
	"fmt"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// GroupResponse wraps the types.QueryGroupResponse to provide additional helper methods.
type GroupResponse struct {
	types.QueryGroupResponse
}

// NewGroupResponse creates a new instance of GroupResponse.
func NewGroupResponse(gr *types.QueryGroupResponse) *GroupResponse {
	return &GroupResponse{*gr}
}

// GetRound1Data retrieves the Round1Commitment for the specified member ID.
func (gr *GroupResponse) GetRound1Data(mid tss.MemberID) (types.Round1Data, error) {
	if int(mid) > len(gr.AllRound1Data) {
		return types.Round1Data{}, fmt.Errorf("No MemberID(%d) in the group", mid)
	}

	data := gr.AllRound1Data[uint64(mid)-1]
	if data == nil {
		return types.Round1Data{}, fmt.Errorf("No Round1Data from MemberID(%d)", mid)
	}

	return *data, nil
}
