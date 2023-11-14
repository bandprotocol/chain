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
func (gr GroupResponse) GetRound1Info(mid tss.MemberID) (types.Round1Info, error) {
	for _, info := range gr.Round1Infos {
		if info.MemberID == mid {
			return info, nil
		}
	}

	return types.Round1Info{}, fmt.Errorf("No Round1Info from MemberID(%d)", mid)
}

// GetRound2Info retrieves the Round1Commitment for the specified member ID.
func (gr GroupResponse) GetRound2Info(mid tss.MemberID) (types.Round2Info, error) {
	for _, info := range gr.Round2Infos {
		if info.MemberID == mid {
			return info, nil
		}
	}

	return types.Round2Info{}, fmt.Errorf("No Round2Info from MemberID(%d)", mid)
}

// GetEncryptedSecretShare retrieves the encrypted secret share from member (Sender) to member (Receiver).
func (gr GroupResponse) GetEncryptedSecretShare(senderID, receiverID tss.MemberID) (tss.EncSecretShare, error) {
	r2Sender, err := gr.GetRound2Info(senderID)
	if err != nil {
		return nil, err
	}

	// Determine which slot of encrypted secret shares is for Receiver
	slot := types.FindMemberSlot(senderID, receiverID)

	// Return error if the slot exceeds length of shares
	if int(slot) >= len(r2Sender.EncryptedSecretShares) {
		return nil, fmt.Errorf("No encrypted secret share from MemberID(%d) to MemberID(%d)", senderID, receiverID)
	}

	return r2Sender.EncryptedSecretShares[slot], nil
}

// GetMemberID returns member's id of the address in the group.
func (gr GroupResponse) GetMemberID(address string) (tss.MemberID, error) {
	for _, member := range gr.Members {
		if member.Address == address {
			return member.ID, nil
		}
	}

	return 0, fmt.Errorf("%s is not the member", address)
}

// IsMember returns boolean to show if the address is the member in the group.
func (gr GroupResponse) IsMember(address string) bool {
	_, err := gr.GetMemberID(address)
	if err != nil {
		return false
	}

	return true
}
