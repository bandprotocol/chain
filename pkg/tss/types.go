package tss

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/bandprotocol/chain/v2/pkg/tss/internal/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// GroupID represents the ID of a group.
type GroupID uint64

// MemberID represents the ID of a member.
type MemberID uint64

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
		return nil, errors.New("length is not 32")
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
	for _, s := range ss {
		scalar, err := s.Parse()
		if err != nil {
			return nil, err
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
	return Point(ParsePublicKey(point))
}

// Parse converts a Point back to a secp256k1.JacobianPoint.
func (p Point) Parse() (*secp256k1.JacobianPoint, error) {
	point, err := PublicKey(p).Point()
	if err != nil {
		return nil, err
	}

	return point, nil
}

// Parse converts a slice of Points into a slice of secp256k1.JacobianPoint.
func (ps Points) Parse() ([]*secp256k1.JacobianPoint, error) {
	var points []*secp256k1.JacobianPoint
	for _, p := range ps {
		point, err := p.Parse()
		if err != nil {
			return nil, err
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

// ParsePrivateKey parses a secp256k1.ModNScalar into a PrivateKey.
func ParsePrivateKey(scalar *secp256k1.ModNScalar) PrivateKey {
	bytes := secp256k1.NewPrivateKey(scalar).Serialize()
	return PrivateKey(bytes)
}

// Parse converts a PrivateKey back to a secp256k1.PrivateKey.
func (pk PrivateKey) Parse() (*secp256k1.PrivateKey, error) {
	if len(pk) != 32 {
		return nil, errors.New("length is not 32")
	}

	return secp256k1.PrivKeyFromBytes(pk), nil
}

// Scalar converts a PrivateKey to a secp256k1.ModNScalar.
func (pk PrivateKey) Scalar() (*secp256k1.ModNScalar, error) {
	privKey, err := pk.Parse()
	if err != nil {
		return nil, err
	}

	return &privKey.Key, nil
}

// PublicKey converts a PrivateKey to a PublicKey.
func (pk PrivateKey) PublicKey() (PublicKey, error) {
	privKey, err := pk.Parse()
	if err != nil {
		return nil, err
	}

	return privKey.PubKey().SerializeCompressed(), nil
}

// Parse converts a slice of PrivateKeys into a slice of secp256k1.PrivateKey.
func (pks PrivateKeys) Parse() ([]*secp256k1.PrivateKey, error) {
	var privKeys []*secp256k1.PrivateKey
	for _, pk := range pks {
		privKey, err := pk.Parse()
		if err != nil {
			return nil, err
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

// ParsePublicKey parses a secp256k1.JacobianPoint into a PublicKey.
func ParsePublicKey(point *secp256k1.JacobianPoint) PublicKey {
	affinePoint := *point
	affinePoint.ToAffine()

	bytes := secp256k1.NewPublicKey(&affinePoint.X, &affinePoint.Y).SerializeCompressed()
	return PublicKey(bytes)
}

// Parse converts a PublicKey back to a secp256k1.PublicKey.
func (pk PublicKey) Parse() (*secp256k1.PublicKey, error) {
	pubKey, err := secp256k1.ParsePubKey(pk)
	if err != nil {
		return nil, err
	}

	return pubKey, nil
}

// Point converts a PublicKey to a secp256k1.JacobianPoint.
func (pk PublicKey) Point() (*secp256k1.JacobianPoint, error) {
	pubKey, err := pk.Parse()
	if err != nil {
		return nil, err
	}

	var point secp256k1.JacobianPoint
	pubKey.AsJacobian(&point)

	return &point, nil
}

// Parse converts a slice of PublicKeys into a slice of secp256k1.PublicKey.
func (pks PublicKeys) Parse() ([]*secp256k1.PublicKey, error) {
	var pubKeys []*secp256k1.PublicKey
	for _, pk := range pks {
		pubKey, err := pk.Parse()
		if err != nil {
			return nil, err
		}

		pubKeys = append(pubKeys, pubKey)
	}

	return pubKeys, nil
}

// Signature represents a signature (r, s) stored as bytes.
// It uses schnorr.Signature as a base implementation for serialization and parsing.
type Signature []byte

// Parse converts a Signature to a schnorr.Signature.
func (s Signature) Parse() (*schnorr.Signature, error) {
	sig, err := schnorr.ParseSignature(s)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

// KeyPair represents a key pair consisting of a private key and a public key.
type KeyPair struct {
	PrivateKey PrivateKey
	PublicKey  PublicKey
}

// KeyPairs represents a slice of KeyPair values.
type KeyPairs []KeyPair
