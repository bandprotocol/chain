package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v3/pkg/tss"
)

// NewCurrentGroup creates a new current group object.
func NewCurrentGroup(id tss.GroupID, activeTime time.Time) CurrentGroup {
	return CurrentGroup{
		GroupID:    id,
		ActiveTime: activeTime,
	}
}

// NewGroupTransition creates a transition object.
func NewGroupTransition(
	signingID tss.SigningID,
	currentGroupID tss.GroupID,
	incomingGroupID tss.GroupID,
	currentGroupPubKey tss.Point,
	incomingGroupPubKey tss.Point,
	status TransitionStatus,
	execTime time.Time,
	isForceTransition bool,
) GroupTransition {
	return GroupTransition{
		SigningID:           signingID,
		CurrentGroupID:      currentGroupID,
		IncomingGroupID:     incomingGroupID,
		CurrentGroupPubKey:  currentGroupPubKey,
		IncomingGroupPubKey: incomingGroupPubKey,
		Status:              status,
		ExecTime:            execTime,
		IsForceTransition:   isForceTransition,
	}
}

// NewMember creates a new member object.
func NewMember(
	address sdk.AccAddress,
	groupID tss.GroupID,
	isActive bool,
	since time.Time,
) Member {
	return Member{
		Address:  address.String(),
		GroupID:  groupID,
		IsActive: isActive,
		Since:    since,
	}
}

// Validate performs basic validation of member information.
func (m Member) Validate() error {
	if _, err := sdk.AccAddressFromBech32(m.Address); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid member address: %s", err)
	}

	if m.GroupID == 0 {
		return ErrInvalidGroupID.Wrap("group id is 0")
	}

	return nil
}

// NewSigning creates a new signing object.
func NewSigning(
	id SigningID,
	feePerSigner sdk.Coins,
	requester sdk.AccAddress,
	currentGroupSigningID tss.SigningID,
	incomingGroupSigningID tss.SigningID,
) Signing {
	return Signing{
		ID:                     id,
		FeePerSigner:           feePerSigner,
		Requester:              requester.String(),
		CurrentGroupSigningID:  currentGroupSigningID,
		IncomingGroupSigningID: incomingGroupSigningID,
	}
}
