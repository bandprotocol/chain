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
	var signature Signature
	signature.R.Set(r)
	signature.S.Set(s)
	return &signature
}

// Serialize returns the Schnorr signature in the more strict format.
//
// The signatures are encoded as:
//
//	bytes at 0-32  jacobian point R with z as 1, encoded by SerializeCompressed of secp256k1.PublicKey
//	bytes at 33-64 s, encoded also as big-endian uint256
func (signature Signature) Serialize() []byte {
	// Total length of returned signature is the length of r and s.
	var b [SignatureSize]byte
	// Make z = 1
	signature.R.ToAffine()
	// Copy compressed bytes of R to first 33 bytes
	pubKey := secp256k1.NewPublicKey(&signature.R.X, &signature.R.Y).SerializeCompressed()
	copy(b[0:33], pubKey)
	// Copy bytes of S 32 bytes after
	signature.S.PutBytesUnchecked(b[33:65])
	return b[:]
}

// ParseSignature parses a signature according to the EC-Schnorr-DCRv0
// specification and enforces the following additional restrictions specific to
// secp256k1:
//
// - The r component must be in the valid range for secp256k1 field elements
// - The s component must be in the valid range for secp256k1 scalars
func ParseSignature(signature []byte) (*Signature, error) {
	// The signature must be the correct length.
	sigLen := len(signature)
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
	pubKey, err := secp256k1.ParsePubKey(signature[0:33])
	if err != nil {
		str := fmt.Sprintf("invalid signature: r is not valid: %s", err.Error())
		return nil, signatureError(ErrSigRTooBig, str)
	}
	pubKey.AsJacobian(&r)

	var s secp256k1.ModNScalar
	if overflow := s.SetByteSlice(signature[33:65]); overflow {
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
func (signature Signature) IsEqual(otherSig *Signature) bool {
	return signature.R.X.Equals(&otherSig.R.X) && signature.R.Y.Equals(&otherSig.R.Y) && signature.R.Z.Equals(&otherSig.R.Z) &&
		signature.S.Equals(&otherSig.S)
}
