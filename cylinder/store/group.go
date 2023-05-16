package store

import "github.com/bandprotocol/chain/v2/pkg/tss"

// Group represents a TSS group.
type Group struct {
	// Persistent
	MemberID tss.MemberID   // Member ID associated with the group
	PrivKey  tss.PrivateKey // Private key associated with the group

	// Temporary
	Coefficients   tss.Scalars    // Coefficients used in the DKG process of the group
	OneTimePrivKey tss.PrivateKey // One-time private key used in the DKG process of the the group
	KeySyms        tss.PublicKeys // Symmetric keys used in the DKG process of the group
}
