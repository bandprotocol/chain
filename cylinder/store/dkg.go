package store

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
)

// DKG represents DKG information of a TSS group.
type DKG struct {
	GroupID        tss.GroupID  `json:"group_id"`          // Group ID associated with the DKG
	MemberID       tss.MemberID `json:"member_id"`         // Member ID associated with the DKG
	Coefficients   tss.Scalars  `json:"coefficients"`      // Coefficients used in the DKG process of the DKG
	OneTimePrivKey tss.Scalar   `json:"one_time_priv_key"` // One-time private key used in the DKG process of the DKG
}
