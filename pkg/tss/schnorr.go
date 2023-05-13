package tss

import (
	"github.com/bandprotocol/chain/v2/pkg/tss/internal/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func Sign(
	rawPrivKey PrivateKey,
	challenge []byte,
	rawNonce Scalar,
) (Signature, error) {
	privKey := rawPrivKey.Parse()
	privKeyScalar := &privKey.Key

	for iterator := uint32(0); ; iterator++ {
		var nonce *secp256k1.ModNScalar
		if rawNonce != nil {
			nonce = rawNonce.Parse()
		} else {
			nonce = GenerateNonce(
				rawPrivKey,
				Hash(challenge),
				iterator,
			).Parse()
		}

		sig, err := schnorr.Sign(privKeyScalar, nonce, challenge)
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
	challenge []byte,
	rawPubKey PublicKey,
	rawGenerator Point,
	rawOverrideSigR PublicKey,
) error {
	sig, err := rawSignature.Parse()
	if err != nil {
		return err
	}

	pubKey, err := rawPubKey.Parse()
	if err != nil {
		return err
	}

	var overrideSigR *secp256k1.JacobianPoint
	if rawOverrideSigR != nil {
		overrideSigR, err = rawOverrideSigR.Point()
		if err != nil {
			return err
		}
	}

	var generator *secp256k1.JacobianPoint
	if rawGenerator != nil {
		generator, err = rawGenerator.Parse()
		if err != nil {
			return err
		}
	}

	err = schnorr.Verify(sig, challenge, pubKey, generator, overrideSigR)
	return err
}
