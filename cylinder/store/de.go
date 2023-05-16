package store

import "github.com/bandprotocol/chain/v2/pkg/tss"

// DE represents private value (D, E) used in TSS signing process.
type DE struct {
	PrivD tss.PrivateKey // Private key d
	PrivE tss.PrivateKey // Private key e
}
