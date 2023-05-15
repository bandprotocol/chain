package tss

import (
	"encoding/hex"
	"fmt"

	"github.com/bandprotocol/chain/v2/pkg/tss/internal/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// /////////////////////////////////////////////
// ID
// /////////////////////////////////////////////
type GroupID uint64

type MemberID uint64

// /////////////////////////////////////////////
// Scalar
// /////////////////////////////////////////////
type Scalar []byte

func ParseScalar(scalar *secp256k1.ModNScalar) Scalar {
	bytes := scalar.Bytes()
	return Scalar(bytes[:])
}

func (s Scalar) Parse() *secp256k1.ModNScalar {
	var scalar secp256k1.ModNScalar
	scalar.SetByteSlice(s)
	return &scalar
}

type Scalars []Scalar

func (ss Scalars) Parse() []*secp256k1.ModNScalar {
	var scalars []*secp256k1.ModNScalar
	for _, s := range ss {
		scalars = append(scalars, s.Parse())
	}
	return scalars
}

// /////////////////////////////////////////////
// Point
// /////////////////////////////////////////////
type Point []byte
type Points []Point

func ParsePoint(point *secp256k1.JacobianPoint) Point {
	return Point(ParsePublicKey(point))
}

func (p Point) Parse() (*secp256k1.JacobianPoint, error) {
	point, err := PublicKey(p).Point()
	if err != nil {
		return nil, err
	}

	return point, nil
}

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

// /////////////////////////////////////////////
// Public key
// /////////////////////////////////////////////
type PublicKey []byte
type PublicKeys []PublicKey

func ParsePublicKey(point *secp256k1.JacobianPoint) PublicKey {
	affinePoint := *point
	affinePoint.ToAffine()

	bytes := secp256k1.NewPublicKey(&affinePoint.X, &affinePoint.Y).SerializeCompressed()
	return PublicKey(bytes)
}

func (pk PublicKey) Parse() (*secp256k1.PublicKey, error) {
	pubKey, err := secp256k1.ParsePubKey(pk)
	if err != nil {
		return nil, err
	}

	return pubKey, nil
}

func (pk PublicKey) Point() (*secp256k1.JacobianPoint, error) {
	pubKey, err := pk.Parse()
	if err != nil {
		return nil, err
	}

	var point secp256k1.JacobianPoint
	pubKey.AsJacobian(&point)

	return &point, nil
}

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

// /////////////////////////////////////////////
// Private key
// /////////////////////////////////////////////
type PrivateKey []byte
type PrivateKeys []PrivateKey

func ParsePrivateKey(scalar secp256k1.ModNScalar) PrivateKey {
	bytes := secp256k1.NewPrivateKey(&scalar).Serialize()
	return PrivateKey(bytes)
}

func (pk PrivateKey) Parse() *secp256k1.PrivateKey {
	return secp256k1.PrivKeyFromBytes(pk)
}

func (pk PrivateKey) Scalar() *secp256k1.ModNScalar {
	return &pk.Parse().Key
}

func (pks PrivateKeys) Parse() []*secp256k1.PrivateKey {
	var privKeys []*secp256k1.PrivateKey
	for _, pk := range pks {
		privKeys = append(privKeys, pk.Parse())
	}

	return privKeys
}

// /////////////////////////////////////////////
// Signature
// /////////////////////////////////////////////
type Signature []byte

func (s Signature) Parse() (*schnorr.Signature, error) {
	sig, err := schnorr.ParseSignature(s)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

// /////////////////////////////////////////////
// Key pair
// /////////////////////////////////////////////
type KeyPair struct {
	PrivateKey PrivateKey
	PublicKey  PublicKey
}

type KeyPairs []KeyPair
