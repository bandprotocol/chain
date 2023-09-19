package store

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
)

// Group represents a TSS group.
type Group struct {
	GroupPubKey tss.Point    `json:"group_pub_key"` // Public key of the group
	MemberID    tss.MemberID `json:"member_id"`     // Member ID associated with the group
	PrivKey     tss.Scalar   `json:"priv_key"`      // Private key associated with the group
}
