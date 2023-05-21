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

// GetRound2Shares retrieves the Round2Shares for the specified member ID.
func (gr *GroupResponse) GetRound2Shares(mid tss.MemberID) (*types.Round2Share, error) {
	if int(mid) > len(gr.Round2Shares) {
		return nil, fmt.Errorf("No Round2Shares from MemberID(%d)", mid)
	}

	return &gr.Round2Shares[mid-1], nil
}

// GetEncryptedSecretShare retrieves the encrypted secret share between specific member I and member J.
func (gr *GroupResponse) GetEncryptedSecretShare(j, i tss.MemberID) (tss.Scalar, error) {
	round2SharesJ, err := gr.GetRound2Shares(j)
	if err != nil {
		return nil, err
	}

	idx := i
	if i > j {
		idx--
	}

	if int(idx) > len(round2SharesJ.EncryptedSecretShares) {
		return nil, fmt.Errorf("No Round2Shares from MemberID(%d) for MemberID(%d)", j, i)
	}

	return round2SharesJ.EncryptedSecretShares[idx-1], nil
}
