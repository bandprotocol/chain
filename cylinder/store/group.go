package store

import "github.com/bandprotocol/chain/v2/pkg/tss"

type Group struct {
	// persistent
	MemberID tss.MemberID
	PrivKey  tss.PrivateKey

	// temporary
	Coefficients   tss.Scalars
	OneTimePrivKey tss.PrivateKey
	KeySyms        tss.PublicKeys
}
