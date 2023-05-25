package tss

import (
	"github.com/bandprotocol/chain/v2/pkg/tss/internal/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Sign generates a schnorr signature for the given private key, challenge, and nonce.
// It returns the signature and an error if the signing process fails.
func Sign(
	rawPrivKey PrivateKey,
	challenge []byte,
	rawNonce Scalar,
) (Signature, error) {
	privKey, err := rawPrivKey.Parse()
	if err != nil {
		return nil, err
	}

	privKeyScalar := &privKey.Key

	for iterator := uint32(0); ; iterator++ {
		// generate nonce if there is no nonce from input parameter
		var nonce *secp256k1.ModNScalar
		if rawNonce != nil {
			nonce, err = rawNonce.Parse()
		} else {
			nonce, err = GenerateNonce(rawPrivKey, Hash(challenge), iterator).Parse()
		}

		if err != nil {
			return nil, err
		}

		var sigR secp256k1.JacobianPoint
		secp256k1.ScalarBaseMultNonConst(nonce, &sigR)
		hash := Hash(ParsePoint(&sigR), challenge)

		sigS, err := schnorr.ComputeSigS(privKeyScalar, nonce, hash)
		nonce.Zero()

		if err != nil {
			// - if there is nonce from input, return error
			// - if not, retry signing with new random nonce
			if rawNonce == nil {
				continue
			}
			return nil, err
		}

		sig := schnorr.NewSignature(&sigR, sigS)
		return sig.Serialize(), nil
	}
}

// Verify verifies the given schnorr signature against the provided challenge, public key, generator point,
// and optional override signature R value.
// It returns an error if the verification process fails.
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

	sigR := &sig.R
	if rawOverrideSigR != nil {
		sigR, err = rawOverrideSigR.Point()
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

	hash := Hash(ParsePoint(&sig.R), challenge)
	return schnorr.Verify(sigR, &sig.S, hash, pubKey, generator)
}
