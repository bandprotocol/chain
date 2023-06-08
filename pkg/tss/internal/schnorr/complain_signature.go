package schnorr

import (
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

const (
	// ComplainSignatureSize is the size of an encoded complain signature.
	ComplainSignatureSize = 98
)

// ComplainSignature is a type representing a complain signature.
type ComplainSignature struct {
	A1 secp256k1.JacobianPoint
	A2 secp256k1.JacobianPoint
	Z  secp256k1.ModNScalar
}

// NewComplainSignature instantiates a new complain signature given some a1, a2 and z values.
func NewComplainSignature(
	a1 *secp256k1.JacobianPoint,
	a2 *secp256k1.JacobianPoint,
	z *secp256k1.ModNScalar,
) *ComplainSignature {
	var sig ComplainSignature
	sig.A1.Set(a1)
	sig.A2.Set(a2)
	sig.Z.Set(z)
	return &sig
}

// Serialize returns the complain signature in the more strict format.
//
// The signatures are encoded as:
//
//	sig[0:33]  jacobian point R with z as 1 (A1), encoded by SerializeCompressed of secp256k1.PublicKey
//	sig[33:66]  jacobian point R with z as 1 (A2), encoded by SerializeCompressed of secp256k1.PublicKey
//	sig[66:98] s, encoded also as big-endian uint256 (Z)
func (sig ComplainSignature) Serialize() []byte {
	// Total length of returned signature is the length of a1, a2 and z.
	var b [ComplainSignatureSize]byte
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

// ParseComplainSignature parses a signature from bytes
//
// - The a1 component must be in the valid range for secp256k1 field elements
// - The a2 component must be in the valid range for secp256k1 field elements
// - The s component must be in the valid range for secp256k1 scalars
func ParseComplainSignature(sig []byte) (*ComplainSignature, error) {
	// The signature must be the correct length.
	sigLen := len(sig)
	if sigLen < ComplainSignatureSize {
		str := fmt.Sprintf("malformed complain signature: too short: %d < %d", sigLen,
			ComplainSignatureSize)
		return nil, signatureError(ErrSigTooShort, str)
	}
	if sigLen > ComplainSignatureSize {
		str := fmt.Sprintf("malformed complain signature: too long: %d > %d", sigLen,
			ComplainSignatureSize)
		return nil, signatureError(ErrSigTooLong, str)
	}

	// The signature is validly encoded at this point, however, enforce
	// additional restrictions to ensure a1 and a2 are the valid jacobian point, and z is in
	// the range [0, n-1] since valid complain signatures are required to be in
	// that range per spec.
	var a1 secp256k1.JacobianPoint
	pubKey, err := secp256k1.ParsePubKey(sig[0:33])
	if err != nil {
		str := fmt.Sprintf("invalid complain signature: a1 is not valid: %s", err.Error())
		return nil, signatureError(ErrSigA1TooBig, str)
	}
	pubKey.AsJacobian(&a1)

	var a2 secp256k1.JacobianPoint
	pubKey, err = secp256k1.ParsePubKey(sig[33:66])
	if err != nil {
		str := fmt.Sprintf("invalid complain signature: a2 is not valid: %s", err.Error())
		return nil, signatureError(ErrSigA2TooBig, str)
	}
	pubKey.AsJacobian(&a2)

	var z secp256k1.ModNScalar
	if overflow := z.SetByteSlice(sig[66:98]); overflow {
		str := "invalid complain signature: z >= group order"
		return nil, signatureError(ErrSigZTooBig, str)
	}

	// Return the complain signature.
	return NewComplainSignature(&a1, &a2, &z), nil
}
