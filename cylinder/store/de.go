package store

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// DE represents private value (D, E) used in TSS signing process.
type DE struct {
	PubDE types.DE   `json:"pub_d"`
	PrivD tss.Scalar `json:"priv_d"` // Private key d
	PrivE tss.Scalar `json:"priv_e"` // Private key e
}
