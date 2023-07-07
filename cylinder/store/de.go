package store

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// DE represents private value (D, E) used in TSS signing process.
type DE struct {
	PrivD tss.Scalar // Private key d
	PrivE tss.Scalar // Private key e
}

func (de DE) PubDE() types.DE {
	return types.DE{
		PubD: de.PrivD.Point(),
		PubE: de.PrivE.Point(),
	}
}
