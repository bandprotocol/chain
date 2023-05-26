package tss

import (
	"errors"

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
		return nil, err
	}

	nonce, err := rawNonce.Parse()
	if err != nil {
		return nil, err
	}

	var sigR secp256k1.JacobianPoint
	secp256k1.ScalarBaseMultNonConst(nonce, &sigR)

	var challenge secp256k1.ModNScalar
	challenge.SetByteSlice(Hash(msg))

	if rawLagrange != nil {
		lagrange, err := rawLagrange.Parse()
		if err != nil {
			return nil, err
		}
		challenge.Mul(lagrange)
	}

	sigS, err := computeSigS(&privKey.Key, nonce, &challenge)
	if err != nil {
		return nil, err
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
		return err
	}

	sigS, err := rawSigS.Parse()
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

	var challenge secp256k1.ModNScalar
	challenge.SetByteSlice(Hash(msg))

	if rawLagrange != nil {
		lagrange, err := rawLagrange.Parse()
		if err != nil {
			return err
		}
		challenge.Mul(lagrange)
	}

	return verify(sigR, sigS, &challenge, pubKey, generator)
}

// computeSigS generates a S part of schnorr signature over the secp256k1 curve
// for the provided challenge using the given nonce, and private key.
func computeSigS(
	privKey, nonce *secp256k1.ModNScalar,
	challenge *secp256k1.ModNScalar,
) (*secp256k1.ModNScalar, error) {
	// G = curve generator
	// n = curve order
	// d = private key
	// c = challenge (hash of message)
	// R, S = signature
	//
	// 1. Fail if d = 0 or d >= n
	// 2. S = k - h*d mod n
	// 3. Return S

	// Step 1.
	//
	// Fail if d = 0 or d >= n
	if privKey.IsZero() {
		return nil, errors.New("private key is zero")
	}

	// Step 2.
	//
	// s = k - c*d mod n
	c := *challenge
	k := *nonce
	S := new(secp256k1.ModNScalar).Mul2(&c, privKey).Negate().Add(&k)
	k.Zero()

	// Step 3.
	//
	// Return S
	return S, nil
}

// verify attempt to verify the signature for the provided challenge, generator and
// secp256k1 public key and either returns nil if successful or a specific error
// indicating why it failed if not successful.
func verify(
	expectR *secp256k1.JacobianPoint,
	sigS *secp256k1.ModNScalar,
	challenge *secp256k1.ModNScalar,
	pubKey *secp256k1.PublicKey,
	generator *secp256k1.JacobianPoint,
) error {
	// G is default curve generator if generator from input is not specified.
	//
	// 1. Fail if Q is not a point on the curve
	// 2. R = s*G + c*Q
	// 3. Fail if R is the point at infinity
	// 4. Verified if R == expectR

	// Step 1.
	//
	// Fail if Q is not a point on the curve
	if !pubKey.IsOnCurve() {
		return errors.New("pubkey point is not on curve")
	}

	// Step 2.
	//
	// R = s*G + h*Q
	c := *challenge
	var Q, R, sG, eQ secp256k1.JacobianPoint
	pubKey.AsJacobian(&Q)
	if generator == nil {
		secp256k1.ScalarBaseMultNonConst(sigS, &sG)
	} else {
		secp256k1.ScalarMultNonConst(sigS, generator, &sG)
	}
	secp256k1.ScalarMultNonConst(&c, &Q, &eQ)
	secp256k1.AddNonConst(&sG, &eQ, &R)

	// Step 3.
	//
	// Fail if R is the point at infinity
	if (R.X.IsZero() && R.Y.IsZero()) || R.Z.IsZero() {
		return errors.New("calculated R point is the point at infinity")
	}
	R.ToAffine()
	expectR.ToAffine()

	// Step 4.
	//
	// Verified if R == expectR
	//
	// Note that R and expectR must be in affine coordinates for this check.
	if !expectR.X.Equals(&R.X) || !expectR.Y.Equals(&R.Y) || !expectR.Z.Equals(&R.Z) {
		return errors.New("calculated R point was not given R")
	}

	return nil
}
