package tss

import (
	"github.com/bandprotocol/chain/v2/pkg/tss/internal/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Sign generates a schnorr signature for the given private key, challenge, and nonce.
// It returns the signature and an error if the signing process fails.
func Sign(
	rawPrivKey Scalar,
	rawChallenge Scalar,
	rawNonce Scalar,
	rawLagrange Scalar,
) (Signature, error) {
	privKey := rawPrivKey.privateKey()
	nonce := rawNonce.modNScalar()
	challenge := rawChallenge.modNScalar()

	var sigR secp256k1.JacobianPoint
	secp256k1.ScalarBaseMultNonConst(nonce, &sigR)

	if rawLagrange != nil {
		lagrange := rawLagrange.modNScalar()
		challenge.Mul(lagrange)
	}

	sigS, err := schnorr.ComputeSigS(privKey, nonce, challenge)
	if err != nil {
		return nil, NewError(err, "compute sig S")
	}

	sig := schnorr.NewSignature(&sigR, sigS)
	return sig.Serialize(), nil
}

// Verify verifies the given schnorr signature against the challenge, public key, generator point,
// and optional override signature R value.
// It returns an error if the verification process fails.
func Verify(
	rawSigR Point,
	rawSigS Scalar,
	rawChallenge Scalar,
	rawPubKey Point,
	rawGenerator Point,
	rawLagrange Scalar,
) error {
	sigR, err := rawSigR.jacobianPoint()
	if err != nil {
		return NewError(err, "parse sig R")
	}

	sigS := rawSigS.modNScalar()

	pubKey, err := rawPubKey.publicKey()
	if err != nil {
		return NewError(err, "parse public key")
	}

	challenge := rawChallenge.modNScalar()

	var generator *secp256k1.JacobianPoint
	if rawGenerator != nil {
		generator, err = rawGenerator.jacobianPoint()
		if err != nil {
			return NewError(err, "parse generator")
		}
	}

	if rawLagrange != nil {
		lagrange := rawLagrange.modNScalar()
		challenge.Mul(lagrange)
	}

	err = schnorr.Verify(sigR, sigS, challenge, pubKey, generator)
	if err != nil {
		return NewError(ErrInvalidSignature, err.Error())
	}

	return nil
}
