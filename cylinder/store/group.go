package store

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
)

// Group represents a TSS group.
type Group struct {
	MemberID tss.MemberID // Member ID associated with the group
	PrivKey  tss.Scalar   // Private key associated with the group
}
