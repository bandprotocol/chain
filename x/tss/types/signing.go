package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

// NewSigning creates a new Signing instance with provided parameters.
func NewSigning(
	gid tss.GroupID,
	assignedMembers []AssignedMember,
	msg []byte,
	groupPubNonce tss.Point,
	signature tss.Signature,
	fee sdk.Coins,
	status SigningStatus,
	requester string,
) Signing {
	return Signing{
		GroupID:         gid,
		AssignedMembers: assignedMembers,
		Message:         msg,
		GroupPubNonce:   groupPubNonce,
		Signature:       signature,
		Fee:             fee,
		Status:          status,
		Requester:       requester,
	}
}
