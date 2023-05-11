package tss

import (
	"github.com/bandprotocol/chain/v2/pkg/tss/internal/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func Sign(
	rawPrivKey PrivateKey,
	commitment []byte,
	rawGenerator Point,
	rawNonce Scalar,
) (Signature, error) {
	privKey := rawPrivKey.Parse()
	privKeyScalar := &privKey.Key

	var generator *secp256k1.JacobianPoint
	if rawGenerator != nil {
		var err error
		generator, err = rawGenerator.Parse()
		if err != nil {
			return nil, err
		}
	}

	for iterator := uint64(0); ; iterator++ {
		nonce := secp256k1.NonceRFC6979(
			rawPrivKey,
			commitment,
			schnorr.RFC6979ExtraDataV0[:],
			nil,
			uint32(iterator),
		)
		if rawNonce != nil {
			nonce = rawNonce.Parse()
		}

		sig, err := schnorr.Sign(privKeyScalar, nonce, commitment, generator)
		nonce.Zero()
		if err != nil {
			if rawNonce == nil {
				continue
			}
			return nil, err
		}

		return sig.Serialize(), nil
	}
}

func Verify(
	rawSignature Signature,
	commitment []byte,
	rawPubKey PublicKey,
	rawGenerator Point,
) error {
	sig, err := rawSignature.Parse()
	if err != nil {
		return err
	}

	pubKey, err := rawPubKey.Parse()
	if err != nil {
		return err
	}

	var generator *secp256k1.JacobianPoint
	if rawGenerator != nil {
		generator, err = rawGenerator.Parse()
		if err != nil {
			return err
		}
	}

	err = schnorr.Verify(sig, commitment, pubKey, generator)
	return err
}
