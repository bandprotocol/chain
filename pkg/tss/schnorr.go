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
	// Serialize input to internal types
	privKey := rawPrivKey.privateKey()
	nonce := rawNonce.modNScalar()
	challenge := rawChallenge.modNScalar()

	// Get signature R from nonce (R = kG)
	var signatureR secp256k1.JacobianPoint
	secp256k1.ScalarBaseMultNonConst(nonce, &signatureR)

	// If there is lagrange value, multiply it into challenge
	// C = CL
	if rawLagrange != nil {
		lagrange := rawLagrange.modNScalar()
		challenge.Mul(lagrange)
	}

	// Compute signature S
	signatureS, err := schnorr.ComputeSignatureS(privKey, nonce, challenge)
	if err != nil {
		return nil, NewError(err, "compute signature S")
	}

	// Construct signature from R and S value
	signature := schnorr.NewSignature(&signatureR, signatureS)

	// Serialize the result to external type
	return signature.Serialize(), nil
}

// Verify verifies the given schnorr signature against the challenge, public key, generator point,
// and optional override signature R value.
// It returns an error if the verification process fails.
func Verify(
	rawSignatureR Point,
	rawSignatureS Scalar,
	rawChallenge Scalar,
	rawPubKey Point,
	rawGenerator Point,
	rawLagrange Scalar,
) error {
	// Serialize input to internal types
	signatureR, err := rawSignatureR.jacobianPoint()
	if err != nil {
		return NewError(err, "parse signature R")
	}

	signatureS := rawSignatureS.modNScalar()

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

	// If there is lagrange value, multiply it into challenge
	// C = CL
	if rawLagrange != nil {
		lagrange := rawLagrange.modNScalar()
		challenge.Mul(lagrange)
	}

	// Verify signature
	err = schnorr.Verify(signatureR, signatureS, challenge, pubKey, generator)
	if err != nil {
		return NewError(ErrInvalidSignature, err.Error())
	}

	return nil
}
