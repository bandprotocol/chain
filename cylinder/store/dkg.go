package store

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
)

// DKG represents DKG information of a TSS group.
type DKG struct {
	MemberID       tss.MemberID // Member ID associated with the group
	Coefficients   tss.Scalars  // Coefficients used in the DKG process of the group
	OneTimePrivKey tss.Scalar   // One-time private key used in the DKG process of the the group
}
