package types

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
)

func NewSigningAttempt(
	signingID tss.SigningID,
	attempt uint64,
	expiredHeight uint64,
	assignedMembers []AssignedMember,
) SigningAttempt {
	return SigningAttempt{
		SigningID:       signingID,
		Attempt:         attempt,
		ExpiredHeight:   expiredHeight,
		AssignedMembers: assignedMembers,
	}
}
