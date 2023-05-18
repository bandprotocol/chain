// Copyright (c) 2013-2014 The btcsuite developers
// Copyright (c) 2015-2022 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package schnorr

import (
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

const (
	// SignatureSize is the size of an encoded Schnorr signature.
	SignatureSize = 65
)

var (
	// rfc6979ExtraDataV0 is the extra data to feed to RFC6979 when generating
	// the deterministic nonce for the EC-Schnorr-DCRv0 scheme.  This ensures
	// the same nonce is not generated for the same message and key as for other
	// signing algorithms such as ECDSA.
	//
	// It is equal to BLAKE-256([]byte("EC-Schnorr-DCRv0")).
	RFC6979ExtraDataV0 = [32]byte{
		0x0b, 0x75, 0xf9, 0x7b, 0x60, 0xe8, 0xa5, 0x76,
		0x28, 0x76, 0xc0, 0x04, 0x82, 0x9e, 0xe9, 0xb9,
		0x26, 0xfa, 0x6f, 0x0d, 0x2e, 0xea, 0xec, 0x3a,
		0x4f, 0xd1, 0x44, 0x6a, 0x76, 0x83, 0x31, 0xcb,
	}
)

// Signature is a type representing a Schnorr signature.
type Signature struct {
	R secp256k1.JacobianPoint
	S secp256k1.ModNScalar
}

// NewSignature instantiates a new signature given some r and s values.
func NewSignature(r *secp256k1.JacobianPoint, s *secp256k1.ModNScalar) *Signature {
	var sig Signature
	sig.R.Set(r)
	sig.S.Set(s)
	return &sig
}

// Serialize returns the Schnorr signature in the more strict format.
//
// The signatures are encoded as:
//
//	sig[0:33]  jacobian point R with z as 1, encoded by SerializeCompressed of secp256k1.PublicKey
//	sig[33:65] s, encoded also as big-endian uint256
func (sig Signature) Serialize() []byte {
	// Total length of returned signature is the length of r and s.
	var b [SignatureSize]byte
	// Make z = 1
	sig.R.ToAffine()
	// Copy compressed bytes of R to first 33 bytes
	pubKey := secp256k1.NewPublicKey(&sig.R.X, &sig.R.Y).SerializeCompressed()
	copy(b[0:33], pubKey)
	// Copy bytes of S 32 bytes after
	sig.S.PutBytesUnchecked(b[33:65])
	return b[:]
}

// ParseSignature parses a signature according to the EC-Schnorr-DCRv0
// specification and enforces the following additional restrictions specific to
// secp256k1:
//
// - The r component must be in the valid range for secp256k1 field elements
// - The s component must be in the valid range for secp256k1 scalars
func ParseSignature(sig []byte) (*Signature, error) {
	// The signature must be the correct length.
	sigLen := len(sig)
	if sigLen < SignatureSize {
		str := fmt.Sprintf("malformed signature: too short: %d < %d", sigLen,
			SignatureSize)
		return nil, signatureError(ErrSigTooShort, str)
	}
	if sigLen > SignatureSize {
		str := fmt.Sprintf("malformed signature: too long: %d > %d", sigLen,
			SignatureSize)
		return nil, signatureError(ErrSigTooLong, str)
	}

	// The signature is validly encoded at this point, however, enforce
	// additional restrictions to ensure r is the valid jacobian point, and s is in
	// the range [0, n-1] since valid Schnorr signatures are required to be in
	// that range per spec.
	var r secp256k1.JacobianPoint
	pubKey, err := secp256k1.ParsePubKey(sig[0:33])
	if err != nil {
		str := fmt.Sprintf("invalid signature: r is not valid: %s", err.Error())
		return nil, signatureError(ErrSigRTooBig, str)
	}
	pubKey.AsJacobian(&r)

	var s secp256k1.ModNScalar
	if overflow := s.SetByteSlice(sig[33:65]); overflow {
		str := "invalid signature: s >= group order"
		return nil, signatureError(ErrSigSTooBig, str)
	}

	// Return the signature.
	return NewSignature(&r, &s), nil
}

// IsEqual compares this Signature instance to the one passed, returning true
// if both Signatures are equivalent. A signature is equivalent to another, if
// they both have the same scalar value for R and S.
// Note: Both R must be affine coordinate.
func (sig Signature) IsEqual(otherSig *Signature) bool {
	return sig.R.X.Equals(&otherSig.R.X) && sig.R.Y.Equals(&otherSig.R.Y) && sig.R.Z.Equals(&otherSig.R.Z) &&
		sig.S.Equals(&otherSig.S)
}

// Verify attempt to verify the signature for the provided challenge, generator and
// secp256k1 public key and either returns nil if successful or a specific error
// indicating why it failed if not successful.
// Note: expectR must be affine coordinate.
func Verify(
	expectR *secp256k1.JacobianPoint,
	sigS *secp256k1.ModNScalar,
	hash []byte,
	pubKey *secp256k1.PublicKey,
	generator *secp256k1.JacobianPoint,
) error {
	// The algorithm for producing a EC-Schnorr-DCRv0 signature is described in
	// README.md and is reproduced here for reference:
	//
	// G is default curve generator if generator from input is not specified.
	//
	// 1. Fail if Q is not a point on the curve
	// 2. Fail if h >= n
	// 3. R = s*G + h*Q
	// 4. Fail if R is the point at infinity
	// 5. Verified if R == expectR

	// Step 1.
	//
	// Fail if Q is not a point on the curve
	if !pubKey.IsOnCurve() {
		str := "pubkey point is not on curve"
		return signatureError(ErrPubKeyNotOnCurve, str)
	}

	// Step 2.
	//
	// Fail if h >= n
	var h secp256k1.ModNScalar
	if overflow := h.SetByteSlice(hash); overflow {
		str := "hash of (R || m) too big"
		return signatureError(ErrSchnorrHashValue, str)
	}

	// Step 3.
	//
	// R = s*G + h*Q
	var Q, R, sG, eQ secp256k1.JacobianPoint
	pubKey.AsJacobian(&Q)
	if generator == nil {
		secp256k1.ScalarBaseMultNonConst(sigS, &sG)
	} else {
		secp256k1.ScalarMultNonConst(sigS, generator, &sG)
	}
	secp256k1.ScalarMultNonConst(&h, &Q, &eQ)
	secp256k1.AddNonConst(&sG, &eQ, &R)

	// Step 4.
	//
	// Fail if R is the point at infinity
	if (R.X.IsZero() && R.Y.IsZero()) || R.Z.IsZero() {
		str := "calculated R point is the point at infinity"
		return signatureError(ErrSigRNotOnCurve, str)
	}
	R.ToAffine()
	expectR.ToAffine()

	// Step 5.
	//
	// Verified if R == expectR
	//
	// Note that R and expectR must be in affine coordinates for this check.
	if !expectR.X.Equals(&R.X) || !expectR.Y.Equals(&R.Y) || !expectR.Z.Equals(&R.Z) {
		str := "calculated R point was not given R"
		return signatureError(ErrUnequalRValues, str)
	}

	return nil
}

// ComputeSigS generates an EC-Schnorr-DCRv0 signature over the secp256k1 curve
// for the provided hash using the given nonce, and private key.
func ComputeSigS(privKey, nonce *secp256k1.ModNScalar, hash []byte) (*secp256k1.ModNScalar, error) {
	// The algorithm for producing a EC-Schnorr-DCRv0 signature is described in
	// README.md and is reproduced here for reference:
	//
	// G = curve generator
	// n = curve order
	// d = private key
	// h = hash of message
	// R, S = signature
	//
	// 1. Fail if d = 0 or d >= n
	// 4. Fail if h >= n
	// 5. S = k - h*d mod n
	// 6. Return S

	// Step 1.
	//
	// Fail if d = 0 or d >= n
	if privKey.IsZero() {
		str := "private key is zero"
		return nil, signatureError(ErrPrivateKeyIsZero, str)
	}

	// Step 2.
	//
	// Fail if h >= N
	var h secp256k1.ModNScalar
	if overflow := h.SetByteSlice(hash); overflow {
		str := "hash of (R || m) too big"
		return nil, signatureError(ErrSchnorrHashValue, str)
	}

	// Step 3.
	//
	// s = k - h*d mod n
	k := *nonce
	S := new(secp256k1.ModNScalar).Mul2(&h, privKey).Negate().Add(&k)
	k.Zero()

	// Step 4.
	//
	// Return S
	return S, nil
}
