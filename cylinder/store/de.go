package store

import (
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// DE represents private value (D, E) used in TSS signing process.
type DE struct {
	PubDE types.DE   `json:"pub_d_e"` // Public key of D and E
	PrivD tss.Scalar `json:"priv_d"`  // Private key d
	PrivE tss.Scalar `json:"priv_e"`  // Private key e
}
