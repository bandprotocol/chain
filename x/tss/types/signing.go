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