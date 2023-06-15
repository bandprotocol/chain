package schnorr

import (
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

const (
	// ComplaintSignatureSize is the size of an encoded complaint signature.
	ComplaintSignatureSize = 98
)

// ComplaintSignature is a type representing a complaint signature.
type ComplaintSignature struct {
	A1 secp256k1.JacobianPoint
	A2 secp256k1.JacobianPoint
	Z  secp256k1.ModNScalar
}

// NewComplaintSignature instantiates a new complaint signature given some a1, a2 and z values.
func NewComplaintSignature(
	a1 *secp256k1.JacobianPoint,
	a2 *secp256k1.JacobianPoint,
	z *secp256k1.ModNScalar,
) *ComplaintSignature {
	var sig ComplaintSignature
	sig.A1.Set(a1)
	sig.A2.Set(a2)
	sig.Z.Set(z)
	return &sig
}

// Serialize returns the complaint signature in the more strict format.
//
// The signatures are encoded as:
//
//	bytes at 0-32  jacobian point R with z as 1 (A1), encoded by SerializeCompressed of secp256k1.PublicKey
//	bytes at 33-65  jacobian point R with z as 1 (A2), encoded by SerializeCompressed of secp256k1.PublicKey
//	bytes at 66-97 s, encoded also as big-endian uint256 (Z)
func (sig ComplaintSignature) Serialize() []byte {
	// Total length of returned signature is the length of a1, a2 and z.
	var b [ComplaintSignatureSize]byte
	// Make z = 1
	sig.A1.ToAffine()
	sig.A2.ToAffine()
	// Copy compressed bytes of A1 to first 33 bytes
	bytes := secp256k1.NewPublicKey(&sig.A1.X, &sig.A1.Y).SerializeCompressed()
	copy(b[0:33], bytes)
	// Copy compressed bytes of A2 to next 33 bytes
	bytes = secp256k1.NewPublicKey(&sig.A2.X, &sig.A2.Y).SerializeCompressed()
	copy(b[33:66], bytes)
	// Copy bytes of S 32 bytes after
	sig.Z.PutBytesUnchecked(b[66:98])
	return b[:]
}

// ParseComplaintSignature parses a signature from bytes
//
// - The a1 component must be in the valid range for secp256k1 field elements
// - The a2 component must be in the valid range for secp256k1 field elements
// - The s component must be in the valid range for secp256k1 scalars
func ParseComplaintSignature(sig []byte) (*ComplaintSignature, error) {
	// The signature must be the correct length.
	sigLen := len(sig)
	if sigLen < ComplaintSignatureSize {
		str := fmt.Sprintf("malformed complaint signature: too short: %d < %d", sigLen,
			ComplaintSignatureSize)
		return nil, signatureError(ErrSigTooShort, str)
	}
	if sigLen > ComplaintSignatureSize {
		str := fmt.Sprintf("malformed complaint signature: too long: %d > %d", sigLen,
			ComplaintSignatureSize)
		return nil, signatureError(ErrSigTooLong, str)
	}

	// The signature is validly encoded at this point, however, enforce
	// additional restrictions to ensure a1 and a2 are the valid jacobian point, and z is in
	// the range [0, n-1] since valid complaint signatures are required to be in
	// that range per spec.
	var a1 secp256k1.JacobianPoint
	pubKey, err := secp256k1.ParsePubKey(sig[0:33])
	if err != nil {
		str := fmt.Sprintf("invalid complaint signature: a1 is not valid: %s", err.Error())
		return nil, signatureError(ErrSigA1TooBig, str)
	}
	pubKey.AsJacobian(&a1)

	var a2 secp256k1.JacobianPoint
	pubKey, err = secp256k1.ParsePubKey(sig[33:66])
	if err != nil {
		str := fmt.Sprintf("invalid complaint signature: a2 is not valid: %s", err.Error())
		return nil, signatureError(ErrSigA2TooBig, str)
	}
	pubKey.AsJacobian(&a2)

	var z secp256k1.ModNScalar
	if overflow := z.SetByteSlice(sig[66:98]); overflow {
		str := "invalid complaint signature: z >= group order"
		return nil, signatureError(ErrSigZTooBig, str)
	}

	// Return the complaint signature.
	return NewComplaintSignature(&a1, &a2, &z), nil
}
