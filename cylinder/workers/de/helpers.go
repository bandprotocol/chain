package de

import (
	"errors"

	"github.com/bandprotocol/chain/v2/cylinder/store"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

const (
	MaxDuplicateDEAttempts = 5
)

type DEGetter interface {
	HasDE(de types.DE) bool
}

// GenerateDEs generates n pairs of DE by using secret value as a random factor
func GenerateDEs(n uint64, secret tss.Scalar, db DEGetter) (privDEs []store.DE, err error) {
	privDEs = make([]store.DE, 0, n)
	attempt := 0

	for len(privDEs) < int(n) && attempt < MaxDuplicateDEAttempts {
		privD, err := tss.GenerateSigningNonce(secret)
		if err != nil {
			return nil, err
		}

		privE, err := tss.GenerateSigningNonce(secret)
		if err != nil {
			return nil, err
		}

		pubDE := types.DE{
			PubD: privD.Point(),
			PubE: privE.Point(),
		}

		if db.HasDE(pubDE) {
			attempt += 1
			continue
		}

		privDEs = append(privDEs, store.DE{
			PubDE: pubDE,
			PrivD: privD,
			PrivE: privE,
		})
	}

	// Unlikely to occur, but included for fail-safe measures.
	if attempt >= MaxDuplicateDEAttempts {
		return nil, errors.New("reach maximum generating duplicated DE")
	}

	return privDEs, nil
}
