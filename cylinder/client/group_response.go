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
