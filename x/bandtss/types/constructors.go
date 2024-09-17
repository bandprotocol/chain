package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

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
	lastActive time.Time,
) Member {
	return Member{
		Address:    address.String(),
		GroupID:    groupID,
		IsActive:   isActive,
		Since:      since,
		LastActive: lastActive,
	}
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
