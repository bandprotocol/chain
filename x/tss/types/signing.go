package types

import (
	"time"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

// NewSigning creates a new Signing instance with provided parameters.
func NewSigning(
	id tss.SigningID,
	currentAttempt uint64,
	gid tss.GroupID,
	groupPubKey tss.Point,
	originatorBz []byte,
	msg []byte,
	groupPubNonce tss.Point,
	signature tss.Signature,
	status SigningStatus,
	createdHeight uint64,
	createdTimestamp time.Time,
) Signing {
	return Signing{
		ID:               id,
		CurrentAttempt:   currentAttempt,
		GroupID:          gid,
		GroupPubKey:      groupPubKey,
		Originator:       originatorBz,
		Message:          msg,
		GroupPubNonce:    groupPubNonce,
		Signature:        signature,
		Status:           status,
		CreatedHeight:    createdHeight,
		CreatedTimestamp: createdTimestamp,
	}
}
