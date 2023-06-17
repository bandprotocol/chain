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

// GetRound1Info retrieves the Round1Commitment for the specified member ID.
func (gr *GroupResponse) GetRound1Info(mid tss.MemberID) (types.Round1Info, error) {
	for _, data := range gr.Round1Infos {
		if data.MemberID == mid {
			return data, nil
		}
	}

	return types.Round1Info{}, fmt.Errorf("No Round1Info from MemberID(%d)", mid)
}

// GetRound2Info retrieves the Round1Commitment for the specified member ID.
func (gr *GroupResponse) GetRound2Info(mid tss.MemberID) (types.Round2Info, error) {
	for _, data := range gr.Round2Infos {
		if data.MemberID == mid {
			return data, nil
		}
	}

	return types.Round2Info{}, fmt.Errorf("No Round2Info from MemberID(%d)", mid)
}

// GetEncryptedSecretShare retrieves the encrypted secret share between specific member I and member J.
func (gr *GroupResponse) GetEncryptedSecretShare(j, i tss.MemberID) (tss.Scalar, error) {
	round2InfoJ, err := gr.GetRound2Info(j)
	if err != nil {
		return nil, err
	}

	// Determine which index of encrypted secret shares is for Member I
	// If Member I > Member J, the index should be reduced by 1 (As J doesn't submit its own encrypt secret share)
	idx := i
	if i > j {
		idx--
	}

	if int(idx) > len(round2InfoJ.EncryptedSecretShares) {
		return nil, fmt.Errorf("No Round2Shares from MemberID(%d) for MemberID(%d)", j, i)
	}

	return round2InfoJ.EncryptedSecretShares[idx-1], nil
}

// IsMember returns boolean to show if the address is the member in the group
func (gr *GroupResponse) IsMember(address string) bool {
	for _, member := range gr.Members {
		if member.Address == address {
			return true
		}
	}

	return false
}
