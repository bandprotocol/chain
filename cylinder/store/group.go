package store

import "github.com/bandprotocol/chain/v2/x/tss/types"

type Group struct {
	// persistent
	MemberID types.MemberID
	PrivKey  types.PrivateKey

	// temporary
	Coefficients   types.Coefficients
	OneTimePrivKey types.PrivateKey
	KeySyms        types.PrivateKeys
}
