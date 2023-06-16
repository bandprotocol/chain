package tss

import (
	"encoding/hex"
	"fmt"

	"github.com/bandprotocol/chain/v2/pkg/tss/internal/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// GroupID represents the ID of a group.
type GroupID uint64

// MemberID represents the ID of a member.
type MemberID uint64

// SigningID represents the ID of a signing.
type SigningID uint64

// Scalar represents a scalar value stored as bytes.
// It uses secp256k1.ModNScalar as a base implementation for serialization and parsing.
type Scalar []byte

// ParseScalar parses a secp256k1.ModNScalar into a Scalar.
func ParseScalar(scalar *secp256k1.ModNScalar) Scalar {
	bytes := scalar.Bytes()
	return Scalar(bytes[:])
}

// Parse converts a Scalar back to a secp256k1.ModNScalar.
func (s Scalar) Parse() (*secp256k1.ModNScalar, error) {
	if len(s) != 32 {
		return nil, NewError(ErrInvalidLength, "length: %d != 32", len(s))
	}

	var scalar secp256k1.ModNScalar
	scalar.SetByteSlice(s)
	return &scalar, nil
}

// Scalars represents a slice of Scalar values.
type Scalars []Scalar

// Parse converts a slice of Scalars into a slice of secp256k1.ModNScalar.
func (ss Scalars) Parse() ([]*secp256k1.ModNScalar, error) {
	var scalars []*secp256k1.ModNScalar
	for idx, s := range ss {
		scalar, err := s.Parse()
		if err != nil {
			return nil, NewError(err, "parse index: %d", idx)
		}

		scalars = append(scalars, scalar)
	}
	return scalars, nil
}

// Point represents a point (x, y, z) stored as bytes.
// It uses secp256k1.JacobianPoint and secp256k1.PublicKey as base implementations for serialization and parsing.
type Point []byte

// Points represents a slice of Point values.
type Points []Point

// ParsePoint parses a secp256k1.JacobianPoint into a Point.
func ParsePoint(point *secp256k1.JacobianPoint) Point {
	return Point(ParsePublicKeyFromPoint(point))
}

// Parse converts a Point back to a secp256k1.JacobianPoint.
func (p Point) Parse() (*secp256k1.JacobianPoint, error) {
	point, err := PublicKey(p).Point()
	if err != nil {
		return nil, NewError(err, "parse to jacobian point")
	}

	return point, nil
}

// Parse converts a slice of Points into a slice of secp256k1.JacobianPoint.
func (ps Points) Parse() ([]*secp256k1.JacobianPoint, error) {
	var points []*secp256k1.JacobianPoint
	for idx, p := range ps {
		point, err := p.Parse()
		if err != nil {
			return nil, NewError(err, "parse index: %d", idx)
		}

		points = append(points, point)
	}

	return points, nil
}

// ToString converts a slice of Points to a string representation.
func (ps Points) ToString() string {
	var points string
	l := len(ps)
	for i, p := range ps {
		if i == l-1 {
			points += hex.EncodeToString(p)
		} else {
			points += fmt.Sprintf("%s,", hex.EncodeToString(p))
		}
	}
	return points
}

// PrivateKey represents a private key stored as bytes.
// It uses secp256k1.ModNScalar as a base implementation for serialization and parsing.
type PrivateKey []byte

// PrivateKeys represents a slice of PrivateKey values.
type PrivateKeys []PrivateKey

// ParsePrivateKey parses a secp256k1.PrivateKey into a PrivateKey.
func ParsePrivateKey(privKey *secp256k1.PrivateKey) PrivateKey {
	bytes := privKey.Serialize()
	return PrivateKey(bytes)
}

// ParsePrivateKeyFromScalar parses a secp256k1.ModNScalar into a PrivateKey.
func ParsePrivateKeyFromScalar(scalar *secp256k1.ModNScalar) PrivateKey {
	bytes := secp256k1.NewPrivateKey(scalar).Serialize()
	return PrivateKey(bytes)
}

// Parse converts a PrivateKey back to a secp256k1.PrivateKey.
func (pk PrivateKey) Parse() (*secp256k1.PrivateKey, error) {
	if len(pk) != 32 {
		return nil, NewError(ErrInvalidLength, "length: %d != 32", len(pk))
	}

	return secp256k1.PrivKeyFromBytes(pk), nil
}

// Scalar converts a PrivateKey to a secp256k1.ModNScalar.
func (pk PrivateKey) Scalar() (*secp256k1.ModNScalar, error) {
	privKey, err := pk.Parse()
	if err != nil {
		return nil, NewError(err, "parse private key")
	}

	return &privKey.Key, nil
}

// PublicKey converts a PrivateKey to a PublicKey.
func (pk PrivateKey) PublicKey() (PublicKey, error) {
	privKey, err := pk.Parse()
	if err != nil {
		return nil, NewError(err, "parse private key")
	}

	return privKey.PubKey().SerializeCompressed(), nil
}

// Parse converts a slice of PrivateKeys into a slice of secp256k1.PrivateKey.
func (pks PrivateKeys) Parse() ([]*secp256k1.PrivateKey, error) {
	var privKeys []*secp256k1.PrivateKey
	for idx, pk := range pks {
		privKey, err := pk.Parse()
		if err != nil {
			return nil, NewError(err, "parse index: %d", idx)
		}

		privKeys = append(privKeys, privKey)
	}

	return privKeys, nil
}

// PublicKey represents a public key stored as bytes.
// It uses secp256k1.JacobianPoint as a base implementation for serialization and parsing.
type PublicKey []byte

// PublicKeys represents a slice of PublicKey values.
type PublicKeys []PublicKey

// ParsePublicKey parses a secp256k1.PublicKey into a PublicKey.
func ParsePublicKey(pubKey *secp256k1.PublicKey) PublicKey {
	bytes := pubKey.SerializeCompressed()
	return PublicKey(bytes)
}

// ParsePublicKeyFromPoint parses a secp256k1.JacobianPoint into a PublicKey.
func ParsePublicKeyFromPoint(point *secp256k1.JacobianPoint) PublicKey {
	affinePoint := *point
	affinePoint.ToAffine()

	bytes := secp256k1.NewPublicKey(&affinePoint.X, &affinePoint.Y).SerializeCompressed()
	return PublicKey(bytes)
}

// Parse converts a PublicKey back to a secp256k1.PublicKey.
func (pk PublicKey) Parse() (*secp256k1.PublicKey, error) {
	pubKey, err := secp256k1.ParsePubKey(pk)
	if err != nil {
		return nil, NewError(ErrParseError, err.Error())
	}

	return pubKey, nil
}

// Point converts a PublicKey to a secp256k1.JacobianPoint.
func (pk PublicKey) Point() (*secp256k1.JacobianPoint, error) {
	pubKey, err := pk.Parse()
	if err != nil {
		return nil, NewError(err, "parse public key")
	}

	var point secp256k1.JacobianPoint
	pubKey.AsJacobian(&point)

	return &point, nil
}

// Parse converts a slice of PublicKeys into a slice of secp256k1.PublicKey.
func (pks PublicKeys) Parse() ([]*secp256k1.PublicKey, error) {
	var pubKeys []*secp256k1.PublicKey
	for idx, pk := range pks {
		pubKey, err := pk.Parse()
		if err != nil {
			return nil, NewError(err, "parse index: %d", idx)
		}

		pubKeys = append(pubKeys, pubKey)
	}

	return pubKeys, nil
}

// Parse converts a slice of PublicKeys into a slice of secp256k1.JacobianPoint.
func (pks PublicKeys) Points() ([]*secp256k1.JacobianPoint, error) {
	var points []*secp256k1.JacobianPoint
	for idx, pk := range pks {
		point, err := pk.Point()
		if err != nil {
			return nil, NewError(err, "parse index: %d", idx)
		}

		points = append(points, point)
	}

	return points, nil
}

// Signature represents a signature (r, s) stored as bytes.
// It uses schnorr.Signature as a base implementation for serialization and parsing.
type Signature []byte

// Signatures represents a slice of Signature values.
type Signatures []Signature

// NewSignature generates a signature from Point (R) and Scalar (S).
// It returns a signature and an error, if any.
func NewSignature(rawR Point, rawS Scalar) (Signature, error) {
	r, err := rawR.Parse()
	if err != nil {
		return nil, NewError(err, "parse R")
	}

	s, err := rawS.Parse()
	if err != nil {
		return nil, NewError(err, "parse S")
	}

	return ParseSignature(schnorr.NewSignature(r, s)), nil
}

// ParseSignature parses a schnorr.Signature into a Signature.
func ParseSignature(sig *schnorr.Signature) Signature {
	return sig.Serialize()
}

// Parse converts a Signature to a schnorr.Signature.
func (s Signature) Parse() (*schnorr.Signature, error) {
	sig, err := schnorr.ParseSignature(s)
	if err != nil {
		return nil, NewError(ErrParseError, err.Error())
	}

	return sig, nil
}

// R returns R part of the signature
func (s Signature) R() Point {
	if len(s) < 33 {
		return []byte{}
	}
	return Point(s[0:33])
}

// S returns S part of the signature
func (s Signature) S() Scalar {
	if len(s) < 65 {
		return []byte{}
	}
	return Scalar(s[33:65])
}

// ComplaintSignature represents a signature (a1, a2, z) stored as bytes.
// It uses schnorr.ComplaintSignature as a base implementation for serialization and parsing.
type ComplaintSignature []byte

// Signatures represents a slice of Signature values.
type ComplaintSignatures []ComplaintSignature

// NewComplaintSignature generates a signature from 2 Points (A1, A2) and Scalar (Z).
// It returns a complaint signature and an error, if any.
func NewComplaintSignature(rawA1 Point, rawA2 Point, rawZ Scalar) (ComplaintSignature, error) {
	a1, err := rawA1.Parse()
	if err != nil {
		return nil, NewError(err, "parse A1")
	}

	a2, err := rawA2.Parse()
	if err != nil {
		return nil, NewError(err, "parse A2")
	}

	z, err := rawZ.Parse()
	if err != nil {
		return nil, NewError(err, "parse Z")
	}

	return ParseComplaintSignature(schnorr.NewComplaintSignature(a1, a2, z)), nil
}

// ParseComplaintSignature parses a schnorr.ComplaintSignature into a Signature.
func ParseComplaintSignature(sig *schnorr.ComplaintSignature) ComplaintSignature {
	return sig.Serialize()
}

// Parse converts a ComplaintSignature to a schnorr.ComplaintSignature.
func (cs ComplaintSignature) Parse() (*schnorr.ComplaintSignature, error) {
	sig, err := schnorr.ParseComplaintSignature(cs)
	if err != nil {
		return nil, NewError(ErrParseError, err.Error())
	}

	return sig, nil
}

// A1 returns A1 part of the complaint signature
func (cs ComplaintSignature) A1() Point {
	if len(cs) < 33 {
		return []byte{}
	}
	return Point(cs[0:33])
}

// A2 returns A2 part of the complaint signature
func (cs ComplaintSignature) A2() Point {
	if len(cs) < 66 {
		return []byte{}
	}
	return Point(cs[33:66])
}

// S returns S part of the signature
func (cs ComplaintSignature) Z() Scalar {
	if len(cs) < 98 {
		return []byte{}
	}
	return Scalar(cs[66:98])
}

// KeyPair represents a key pair consisting of a private key and a public key.
type KeyPair struct {
	PrivKey PrivateKey
	PubKey  PublicKey
}

// KeyPairs represents a slice of KeyPair values.
type KeyPairs []KeyPair
