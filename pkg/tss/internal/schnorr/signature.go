// Copyright (c) 2013-2014 The btcsuite developers
// Copyright (c) 2015-2022 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package schnorr

import (
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// SignatureSize is the size of an encoded Schnorr signature.
	SignatureSize = 64

	// scalarSize is the size of an encoded big endian scalar.
	scalarSize = 32
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
	r secp256k1.FieldVal
	s secp256k1.ModNScalar
}

// NewSignature instantiates a new signature given some r and s values.
func NewSignature(r *secp256k1.FieldVal, s *secp256k1.ModNScalar) *Signature {
	var sig Signature
	sig.r.Set(r).Normalize()
	sig.s.Set(s)
	return &sig
}

// Serialize returns the Schnorr signature in the more strict format.
//
// The signatures are encoded as:
//
//	sig[0:32]  x coordinate of the point R, encoded as a big-endian uint256
//	sig[32:64] s, encoded also as big-endian uint256
func (sig Signature) Serialize() []byte {
	// Total length of returned signature is the length of r and s.
	var b [SignatureSize]byte
	sig.r.PutBytesUnchecked(b[0:32])
	sig.s.PutBytesUnchecked(b[32:64])
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
	// additional restrictions to ensure r is in the range [0, p-1], and s is in
	// the range [0, n-1] since valid Schnorr signatures are required to be in
	// that range per spec.
	var r secp256k1.FieldVal
	if overflow := r.SetByteSlice(sig[0:32]); overflow {
		str := "invalid signature: r >= field prime"
		return nil, signatureError(ErrSigRTooBig, str)
	}
	var s secp256k1.ModNScalar
	if overflow := s.SetByteSlice(sig[32:64]); overflow {
		str := "invalid signature: s >= group order"
		return nil, signatureError(ErrSigSTooBig, str)
	}

	// Return the signature.
	return NewSignature(&r, &s), nil
}

// IsEqual compares this Signature instance to the one passed, returning true
// if both Signatures are equivalent. A signature is equivalent to another, if
// they both have the same scalar value for R and S.
func (sig Signature) IsEqual(otherSig *Signature) bool {
	return sig.r.Equals(&otherSig.r) && sig.s.Equals(&otherSig.s)
}

// Verify attempt to verify the signature for the provided data, generator and
// secp256k1 public key and either returns nil if successful or a specific error
// indicating why it failed if not successful.
func Verify(sig *Signature, data []byte, pubKey *secp256k1.PublicKey, generator *secp256k1.JacobianPoint) error {
	// The algorithm for producing a EC-Schnorr-DCRv0 signature is described in
	// README.md and is reproduced here for reference:
	//
	// G is default curve generator if generator from input is not specified.
	//
	// 1. Fail if Q is not a point on the curve
	// 2. Fail if r >= p
	// 3. Fail if s >= n
	// 4. e = Keccak256(r || m) (Ensure r is padded to 32 bytes)
	// 5. Fail if e >= n
	// 6. R = s*G + e*Q
	// 7. Fail if R is the point at infinity
	// 8. Fail if R.y is odd
	// 9. Verified if R.x == r

	// Step 1.
	//
	// Fail if Q is not a point on the curve
	if !pubKey.IsOnCurve() {
		str := "pubkey point is not on curve"
		return signatureError(ErrPubKeyNotOnCurve, str)
	}

	// Step 2.
	//
	// Fail if r >= p
	//
	// Note this is already handled by the fact r is a field element.

	// Step 3.
	//
	// Fail if s >= n
	//
	// Note this is already handled by the fact s is a mod n scalar.

	// Step 4.
	//
	// e = Keccak256(r || m) (Ensure r is padded to 32 bytes)
	var commitmentInput [scalarSize]byte
	sig.r.PutBytesUnchecked(commitmentInput[0:scalarSize])
	commitment := crypto.Keccak256(append(commitmentInput[:], data...))

	// Step 5.
	//
	// Fail if e >= n
	var e secp256k1.ModNScalar
	if overflow := e.SetByteSlice(commitment); overflow {
		str := "hash of (R || m) too big"
		return signatureError(ErrSchnorrHashValue, str)
	}

	// Step 6.
	//
	// R = s*G + e*Q
	var Q, R, sG, eQ secp256k1.JacobianPoint
	pubKey.AsJacobian(&Q)
	if generator == nil {
		secp256k1.ScalarBaseMultNonConst(&sig.s, &sG)
	} else {
		secp256k1.ScalarMultNonConst(&sig.s, generator, &sG)
	}
	secp256k1.ScalarMultNonConst(&e, &Q, &eQ)
	secp256k1.AddNonConst(&sG, &eQ, &R)

	// Step 7.
	//
	// Fail if R is the point at infinity
	if (R.X.IsZero() && R.Y.IsZero()) || R.Z.IsZero() {
		str := "calculated R point is the point at infinity"
		return signatureError(ErrSigRNotOnCurve, str)
	}

	// Step 8.
	//
	// Fail if R.y is odd
	//
	// Note that R must be in affine coordinates for this check.
	R.ToAffine()
	if R.Y.IsOdd() {
		str := "calculated R y-value is odd"
		return signatureError(ErrSigRYIsOdd, str)
	}

	// Step 9.
	//
	// Verified if R.x == r
	//
	// Note that R must be in affine coordinates for this check.
	if !sig.r.Equals(&R.X) {
		str := "calculated R point was not given R"
		return signatureError(ErrUnequalRValues, str)
	}

	return nil
}

// Sign generates an EC-Schnorr-DCRv0 signature over the secp256k1 curve
// for the provided data using the given nonce, private key, and generator.  The produced signature is
// deterministic (same message, nonce, generator, and key yield the same signature) and
// canonical.
func Sign(
	privKey, nonce *secp256k1.ModNScalar,
	data []byte,
	generator *secp256k1.JacobianPoint,
) (*Signature, error) {
	// The algorithm for producing a EC-Schnorr-DCRv0 signature is described in
	// README.md and is reproduced here for reference:
	//
	// G = curve generator
	// P = custom curve generator from input
	// n = curve order
	// d = private key
	// m = message
	// r, s = signature
	//
	// 1. Fail if d = 0 or d >= n
	// 2. R = kG if P is nil else kP
	// 3. Negate nonce k if R.y is odd (R.y is the y coordinate of the point R)
	// 4. r = R.x (R.x is the x coordinate of the point R)
	// 5. e = Keccak256(r || m) (Ensure r is padded to 32 bytes)
	// 6. Repeat from step 3 (with iteration + 1) if e >= n
	// 7. s = k - e*d mod n
	// 8. Return (r, s)

	// Step 1.
	//
	// Fail if d = 0 or d >= n
	if privKey.IsZero() {
		str := "private key is zero"
		return nil, signatureError(ErrPrivateKeyIsZero, str)
	}

	// Step 2.
	//
	// R = kG if P is nil else kP
	var R secp256k1.JacobianPoint
	k := *nonce
	if generator == nil {
		secp256k1.ScalarBaseMultNonConst(&k, &R)
	} else {
		secp256k1.ScalarMultNonConst(&k, generator, &R)
	}

	// Step 3.
	//
	// Negate nonce k if R.y is odd (R.y is the y coordinate of the point R)
	//
	// Note that R must be in affine coordinates for this check.
	R.ToAffine()
	if R.Y.IsOdd() {
		k.Negate()
	}

	// Step 4.
	//
	// r = R.x (R.x is the x coordinate of the point R)
	r := &R.X

	// Step 5.
	//
	// e = Keccak256(r || m) (Ensure r is padded to 32 bytes)
	var commitmentInput [scalarSize]byte
	r.PutBytesUnchecked(commitmentInput[0:scalarSize])
	commitment := crypto.Keccak256(append(commitmentInput[:], data...))

	// Step 6.
	//
	// Repeat from step 1 (with iteration + 1) if e >= N
	var e secp256k1.ModNScalar
	if overflow := e.SetByteSlice(commitment); overflow {
		k.Zero()
		str := "hash of (R || m) too big"
		return nil, signatureError(ErrSchnorrHashValue, str)
	}

	// Step 7.
	//
	// s = k - e*d mod n
	s := new(secp256k1.ModNScalar).Mul2(&e, privKey).Negate().Add(&k)
	k.Zero()

	// Step 8.
	//
	// Return (r, s)
	return NewSignature(r, s), nil
}
