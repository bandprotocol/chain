package tss

import (
	"github.com/bandprotocol/chain/v2/pkg/tss/internal/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Sign generates a schnorr signature for the given private key, message, and nonce.
// It returns the signature and an error if the signing process fails.
func Sign(
	rawPrivKey PrivateKey,
	msg []byte,
	rawNonce Scalar,
	rawLagrange Scalar,
) (Signature, error) {
	privKey, err := rawPrivKey.Parse()
	if err != nil {
		return nil, NewError(err, "parse private key")
	}

	nonce, err := rawNonce.Parse()
	if err != nil {
		return nil, NewError(err, "parse nonce")
	}

	var sigR secp256k1.JacobianPoint
	secp256k1.ScalarBaseMultNonConst(nonce, &sigR)

	var challenge secp256k1.ModNScalar
	challenge.SetByteSlice(Hash(msg))

	if rawLagrange != nil {
		lagrange, err := rawLagrange.Parse()
		if err != nil {
			return nil, NewError(err, "parse lagrange")
		}
		challenge.Mul(lagrange)
	}

	sigS, err := schnorr.ComputeSigS(privKey, nonce, &challenge)
	if err != nil {
		return nil, NewError(err, "compute sig S")
	}

	sig := schnorr.NewSignature(&sigR, sigS)
	return sig.Serialize(), nil
}

// Verify verifies the given schnorr signature against the provided msessage, public key, generator point,
// and optional override signature R value.
// It returns an error if the verification process fails.
func Verify(
	rawSigR Point,
	rawSigS Scalar,
	msg []byte,
	rawPubKey PublicKey,
	rawGenerator Point,
	rawLagrange Scalar,
) error {
	sigR, err := rawSigR.Parse()
	if err != nil {
		return NewError(err, "parse sig R")
	}

	sigS, err := rawSigS.Parse()
	if err != nil {
		return NewError(err, "parse sig S")
	}

	pubKey, err := rawPubKey.Parse()
	if err != nil {
		return NewError(err, "parse public key")
	}

	var generator *secp256k1.JacobianPoint
	if rawGenerator != nil {
		generator, err = rawGenerator.Parse()
		if err != nil {
			return NewError(err, "parse generator")
		}
	}

	var challenge secp256k1.ModNScalar
	challenge.SetByteSlice(Hash(msg))

	if rawLagrange != nil {
		lagrange, err := rawLagrange.Parse()
		if err != nil {
			return NewError(err, "parse lagrange")
		}
		challenge.Mul(lagrange)
	}

	err = schnorr.Verify(sigR, sigS, &challenge, pubKey, generator)
	if err != nil {
		return NewError(ErrInvalidSignature, err.Error())
	}

	return nil
}
