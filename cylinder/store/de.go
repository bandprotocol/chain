package store

import "github.com/bandprotocol/chain/v2/pkg/tss"

// DE represents private value (D, E) used in TSS signing process.
type DE struct {
	PrivD tss.Scalar // Private key d
	PrivE tss.Scalar // Private key e
}
