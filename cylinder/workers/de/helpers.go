package de

import (
	"github.com/bandprotocol/chain/v2/cylinder/store"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// GenerateDEs generates n pairs of DE by using secret value as a random factor
func GenerateDEs(n uint64, secret tss.Scalar) (privDEs []store.DE, err error) {
	for i := uint64(1); i <= n; i++ {
		privD, err := tss.GenerateSigningNonce(secret)
		if err != nil {
			return nil, err
		}

		privE, err := tss.GenerateSigningNonce(secret)
		if err != nil {
			return nil, err
		}

		privDEs = append(privDEs, store.DE{
			PubDE: types.DE{
				PubD: privD.Point(),
				PubE: privE.Point(),
			},
			PrivD: privD,
			PrivE: privE,
		})
	}

	return privDEs, nil
}
