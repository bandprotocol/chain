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
	for _, data := range gr.AllRound1Data {
		if data.MemberID == mid {
			return data, nil
		}
	}

	return types.Round1Data{}, fmt.Errorf("No Round1Data from MemberID(%d)", mid)
}

// GetRound1Data retrieves the Round1Commitment for the specified member ID.
func (gr *GroupResponse) GetRound2Data(mid tss.MemberID) (types.Round2Data, error) {
	if int(mid) > len(gr.AllRound2Data) {
		return types.Round2Data{}, fmt.Errorf("No MemberID(%d) in the group", mid)
	}

	data := gr.AllRound2Data[uint64(mid)-1]
	if data == nil {
		return types.Round2Data{}, fmt.Errorf("No Round2Data from MemberID(%d)", mid)
	}

	return *data, nil
}

// GetEncryptedSecretShare retrieves the encrypted secret share between specific member I and member J.
func (gr *GroupResponse) GetEncryptedSecretShare(j, i tss.MemberID) (tss.Scalar, error) {
	round2DataJ, err := gr.GetRound2Data(j)
	if err != nil {
		return nil, err
	}

	idx := i
	if i > j {
		idx--
	}

	if int(idx) > len(round2DataJ.EncryptedSecretShares) {
		return nil, fmt.Errorf("No Round2Shares from MemberID(%d) for MemberID(%d)", j, i)
	}

	return round2DataJ.EncryptedSecretShares[idx-1], nil
}
