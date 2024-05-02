package types

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
)

// NewSigning creates a new Signing instance with provided parameters.
func NewSigning(
	gid tss.GroupID,
	groupPubKey tss.Point,
	assignedMembers AssignedMembers,
	msg []byte,
	groupPubNonce tss.Point,
	signature tss.Signature,
	status SigningStatus,
) Signing {
	return Signing{
		GroupID:         gid,
		GroupPubKey:     groupPubKey,
		AssignedMembers: assignedMembers,
		Message:         msg,
		GroupPubNonce:   groupPubNonce,
		Signature:       signature,
		Status:          status,
	}
}

// IsFailed check whether the signing is failed due to expired or fail within the execution.
func (s Signing) IsFailed() bool {
	return s.Status == SIGNING_STATUS_EXPIRED || s.Status == SIGNING_STATUS_FALLEN
}
