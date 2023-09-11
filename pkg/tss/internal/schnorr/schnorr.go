package schnorr

import "github.com/decred/dcrd/dcrec/secp256k1/v4"

// ComputeSignatureS generates a S part of schnorr signature over the secp256k1 curve
// for the provided challenge using the given nonce, and private key.
func ComputeSignatureS(
	privKey *secp256k1.PrivateKey, nonce *secp256k1.ModNScalar,
	challenge *secp256k1.ModNScalar,
) (*secp256k1.ModNScalar, error) {
	// G = curve generator
	// n = curve order
	// d = private key
	// c = challenge (hash of message)
	// R, S = signature
	//
	// 1. Fail if d = nil or d = 0
	// 2. S = k + c*d mod n
	// 3. Return S

	// Step 1.
	//
	// Fail if d = nil or d = 0
	if privKey == nil || privKey.Key.IsZero() {
		return nil, ErrPrivateKeyZero
	}

	// Step 2.
	//
	// s = k + c*d mod n
	c := *challenge
	k := *nonce
	S := new(secp256k1.ModNScalar).Mul2(&c, &privKey.Key).Add(&k)
	k.Zero()

	// Step 3.
	//
	// Return S
	return S, nil
}

// Verify attempt to verify the signature for the provided challenge, generator and
// secp256k1 public key and either returns nil if successful or a specific error
// indicating why it failed if not successful.
func Verify(
	expectR *secp256k1.JacobianPoint,
	signatureS *secp256k1.ModNScalar,
	challenge *secp256k1.ModNScalar,
	pubKey *secp256k1.PublicKey,
	generator *secp256k1.JacobianPoint,
) error {
	// G is default curve generator if generator from input is not specified.
	//
	// 1. Fail if Q is not a point on the curve
	// 2. R = s*G - c*Q
	// 3. Fail if R is the point at infinity
	// 4. Verified if R == expectR

	// Step 1.
	//
	// Fail if Q is not a point on the curve
	if !pubKey.IsOnCurve() {
		return ErrNotOnCurve
	}

	// Step 2.
	//
	// R = s*G - c*Q
	c := *challenge
	var Q, R, sG, cQ secp256k1.JacobianPoint
	pubKey.AsJacobian(&Q)
	if generator == nil {
		secp256k1.ScalarBaseMultNonConst(signatureS, &sG)
	} else {
		secp256k1.ScalarMultNonConst(signatureS, generator, &sG)
	}
	secp256k1.ScalarMultNonConst(c.Negate(), &Q, &cQ)
	secp256k1.AddNonConst(&sG, &cQ, &R)

	// Step 3.
	//
	// Fail if R is the point at infinity
	if (R.X.IsZero() && R.Y.IsZero()) || R.Z.IsZero() {
		return ErrRInfinity
	}
	R.ToAffine()
	expectR.ToAffine()

	// Step 4.
	//
	// Verified if R == expectR
	//
	// Note that R and expectR must be in affine coordinates for this check.
	if !expectR.X.Equals(&R.X) || !expectR.Y.Equals(&R.Y) || !expectR.Z.Equals(&R.Z) {
		return ErrIncorrectR
	}

	return nil
}
