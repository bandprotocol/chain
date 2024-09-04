package types

import (
	"time"

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
