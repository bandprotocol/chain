package tss

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/bandprotocol/chain/v2/pkg/tss/internal/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// GroupID represents the ID of a group.
type GroupID uint64

// MemberID represents the ID of a member.
// Please note that the MemberID can only be 1, 2, 3, ..., 2**64 - 1
type MemberID uint64

// NewMemberID creates a new MemberID from any value, ensuring it is within the valid range.
// Panic if the input value cannot be converted to a uint64 or if it is less than 1.
func NewMemberID(value interface{}) MemberID {
	str := fmt.Sprint(value)
	v, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		panic("NewMemberID: conversion to uint64 failed")
	}
	if v < 1 {
		panic("NewMemberID: the value must be greater than 0")
	}
	return MemberID(v)
}

// MemberIDZero returns a MemberID with a value of 0. This is outside the valid range for MemberID values,
// and thus should not be used as a valid MemberID in any operational context.
// It is primarily intended for use in scenarios where a placeholder or default value is needed.
func MemberIDZero() MemberID {
	return MemberID(0)
}

// SigningID represents the ID of a signing.
type SigningID uint64

// Scalar represents a scalar value stored as bytes.
// It uses secp256k1.ModNScalar and secp256k1.PrivateKey as a base implementation for serialization and parsing.
type Scalar []byte

// NewScalar constructs a Scalar from bytes.
func NewScalar(bytes []byte) (Scalar, error) {
	// Create a Scalar from the provided bytes.
	scalar := Scalar(bytes)

	// Check the validity of the scalar value.
	if err := scalar.Validate(); err != nil {
		return nil, NewError(err, "check valid")
	}

	return scalar, nil
}

// NewScalarFromModNScalar parses a secp256k1.ModNScalar into a Scalar.
func NewScalarFromModNScalar(scalar *secp256k1.ModNScalar) Scalar {
	bytes := scalar.Bytes()
	return bytes[:]
}

// NewScalarFromPrivateKey parses a secp256k1.PrivateKey into a Scalar.
func NewScalarFromPrivateKey(privKey *secp256k1.PrivateKey) Scalar {
	return privKey.Serialize()
}

// Marshal needed for protobuf compatibility
func (s Scalar) Marshal() ([]byte, error) {
	return s, nil
}

// Unmarshal needed for protobuf compatibility
func (s *Scalar) Unmarshal(data []byte) error {
	*s = data
	return nil
}

// MarshalJSON converts the Scalar to its JSON representation.
func (s Scalar) MarshalJSON() ([]byte, error) {
	str := strings.ToUpper(hex.EncodeToString(s))
	jbz := make([]byte, len(str)+2)
	jbz[0] = '"'
	copy(jbz[1:], str)
	jbz[len(jbz)-1] = '"'
	return jbz, nil
}

// UnmarshalJSON parses a JSON string into a Scalar.
func (s *Scalar) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid hex string: %s", data)
	}
	bz2, err := hex.DecodeString(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}
	*s = bz2
	return nil
}

// Bytes returns the underlying byte slice of the Scalar.
func (s Scalar) Bytes() []byte {
	return s
}

// String returns the hexadecimal representation of the Scalar in uppercase.
func (s Scalar) String() string {
	return strings.ToUpper(hex.EncodeToString(s))
}

// Validate returns an error if the scalar value is invalid.
func (s Scalar) Validate() error {
	// Check the length of the Scalar value.
	if len(s) != 32 {
		return NewError(ErrInvalidLength, "length: %d != 32", len(s))
	}

	// Set the byte slice value of the Scalar.
	var scalar secp256k1.ModNScalar
	overflow := scalar.SetByteSlice(s)

	// Check if the Scalar is zero or if there was an overflow.
	if scalar.IsZero() || overflow {
		return NewError(ErrInvalidOrder, "set bytes")
	}

	return nil
}

// Point converts a Scalar to a Point.
func (s Scalar) Point() Point {
	point := s.jacobianPoint()
	return NewPointFromJacobianPoint(point)
}

// modNScalar converts a Scalar back to a secp256k1.ModNScalar.
func (s Scalar) modNScalar() *secp256k1.ModNScalar {
	var scalar secp256k1.ModNScalar
	scalar.SetByteSlice(s)

	return &scalar
}

// privateKey converts a Scalar back to a secp256k1.PrivateKey.
func (s Scalar) privateKey() *secp256k1.PrivateKey {
	scalar := s.modNScalar()
	return secp256k1.NewPrivateKey(scalar)
}

// publicKey converts a Scalar back to a secp256k1.PublicKey.
func (s Scalar) publicKey() *secp256k1.PublicKey {
	return s.privateKey().PubKey()
}

// jacobianPoint converts a Scalar to a secp256k1.JacobianPoint.
func (s Scalar) jacobianPoint() *secp256k1.JacobianPoint {
	// Convert the Scalar to a ModNScalar.
	scalar := s.modNScalar()

	// Create a new JacobianPoint by performing scalar base multiplication.
	var point secp256k1.JacobianPoint
	secp256k1.ScalarBaseMultNonConst(scalar, &point)
	point.ToAffine()

	return &point
}

// Scalars represents a slice of Scalar values.
type Scalars []Scalar

// modNScalars converts a slice of Scalars into a slice of secp256k1.ModNScalar.
func (ss Scalars) modNScalars() []*secp256k1.ModNScalar {
	var scalars []*secp256k1.ModNScalar
	for _, s := range ss {
		scalar := s.modNScalar()
		scalars = append(scalars, scalar)
	}

	return scalars
}

// jacobianPoints converts a slice of Scalars into a slice of secp256k1.JacobianPoint.
func (ss Scalars) jacobianPoints() []*secp256k1.JacobianPoint {
	var points []*secp256k1.JacobianPoint
	for _, s := range ss {
		point := s.jacobianPoint()
		points = append(points, point)
	}

	return points
}

// Point represents a point (x, y, z) stored as bytes.
// It uses secp256k1.JacobianPoint and secp256k1.PublicKey as base implementations for serialization and parsing.
type Point []byte

// NewPoint constructs a Point from bytes.
func NewPoint(bytes []byte) (Point, error) {
	// Create a Point from the provided bytes.
	point := Point(bytes)

	// Check the validity of the point value.
	if err := point.Validate(); err != nil {
		return nil, NewError(err, "check valid")
	}

	return point, nil
}

// NewPointFromJacobianPoint parses a secp256k1.JacobianPoint into a Point.
func NewPointFromJacobianPoint(point *secp256k1.JacobianPoint) Point {
	// Convert the JacobianPoint to affine coordinates.
	affinePoint := *point
	affinePoint.ToAffine()

	// Serialize the affine coordinates into bytes and create a Point from it.
	bytes := secp256k1.NewPublicKey(&affinePoint.X, &affinePoint.Y).SerializeCompressed()
	return Point(bytes)
}

// NewPointFromPublicKey parses a secp256k1.PublicKey into a Point.
func NewPointFromPublicKey(pubKey *secp256k1.PublicKey) Point {
	// Serialize the PublicKey into bytes and create a Point from it.
	bytes := pubKey.SerializeCompressed()
	return Point(bytes)
}

// Marshal needed for protobuf compatibility
func (p Point) Marshal() ([]byte, error) {
	return p, nil
}

// Unmarshal needed for protobuf compatibility
func (p *Point) Unmarshal(data []byte) error {
	*p = data
	return nil
}

// MarshalJSON converts the Point to its JSON representation.
func (p Point) MarshalJSON() ([]byte, error) {
	str := strings.ToUpper(hex.EncodeToString(p))
	jbz := make([]byte, len(str)+2)
	jbz[0] = '"'
	copy(jbz[1:], str)
	jbz[len(jbz)-1] = '"'
	return jbz, nil
}

// UnmarshalJSON parses a JSON string into a Point.
func (p *Point) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid hex string: %s", data)
	}
	bz2, err := hex.DecodeString(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}
	*p = bz2
	return nil
}

// Bytes returns the underlying byte slice of the Point.
func (p Point) Bytes() []byte {
	return p
}

// String returns the hexadecimal representation of the Point in uppercase.
func (p Point) String() string {
	return strings.ToUpper(hex.EncodeToString(p))
}

// Validate returns an error if the value is invalid.
func (p Point) Validate() error {
	if _, err := p.publicKey(); err != nil {
		return NewError(err, "check valid")
	}

	return nil
}

// Address returns an ethereum address of the point
func (p Point) Address() ([]byte, error) {
	pubKey, err := p.publicKey()
	if err != nil {
		return nil, err
	}

	return Hash(pubKey.X().Bytes(), pubKey.Y().Bytes())[12:], nil
}

// publicKey converts a Point back to a secp256k1.PublicKey.
func (p Point) publicKey() (*secp256k1.PublicKey, error) {
	pubKey, err := secp256k1.ParsePubKey(p)
	if err != nil {
		return nil, NewError(ErrParseError, err.Error())
	}

	return pubKey, nil
}

// jacobianPoint converts a Point back to a secp256k1.JacobianPoint.
func (p Point) jacobianPoint() (*secp256k1.JacobianPoint, error) {
	pubKey, err := p.publicKey()
	if err != nil {
		return nil, NewError(err, "parse public key")
	}

	var point secp256k1.JacobianPoint
	pubKey.AsJacobian(&point)

	return &point, nil
}

// Points represents a slice of Point values.
type Points []Point

// jacobianPoints converts a slice of Points into a slice of secp256k1.JacobianPoint.
func (ps Points) jacobianPoints() ([]*secp256k1.JacobianPoint, error) {
	var points []*secp256k1.JacobianPoint
	for idx, p := range ps {
		point, err := p.jacobianPoint()
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

// NewSignature constructs a Signature from bytes.
func NewSignature(bytes []byte) (Signature, error) {
	signature := Signature(bytes)
	if err := signature.Validate(); err != nil {
		return nil, NewError(err, "check valid")
	}

	return signature, nil
}

// NewSignatureFromComponents generates a signature from Point (R) and Scalar (S).
// It returns a Signature and an error, if any.
func NewSignatureFromComponents(rawR Point, rawS Scalar) (Signature, error) {
	r, err := rawR.jacobianPoint()
	if err != nil {
		return nil, NewError(err, "parse r")
	}

	s := rawS.modNScalar()

	return NewSignatureFromType(schnorr.NewSignature(r, s)), nil
}

// NewSignatureFromType parses a schnorr.Signature into a Signature.
func NewSignatureFromType(signature *schnorr.Signature) Signature {
	return signature.Serialize()
}

// Marshal needed for protobuf compatibility
func (s Signature) Marshal() ([]byte, error) {
	return s, nil
}

// Unmarshal needed for protobuf compatibility
func (s *Signature) Unmarshal(data []byte) error {
	*s = data
	return nil
}

// MarshalJSON converts the Signature to its JSON representation.
func (s Signature) MarshalJSON() ([]byte, error) {
	str := strings.ToUpper(hex.EncodeToString(s))
	jbz := make([]byte, len(str)+2)
	jbz[0] = '"'
	copy(jbz[1:], str)
	jbz[len(jbz)-1] = '"'
	return jbz, nil
}

// UnmarshalJSON parses a JSON string into a Signature.
func (s *Signature) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid hex string: %s", data)
	}
	bz2, err := hex.DecodeString(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}
	*s = bz2
	return nil
}

// Bytes returns the underlying byte slice of the Signature.
func (s Signature) Bytes() []byte {
	return s
}

// String returns the hexadecimal representation of the Signature in uppercase.
func (s Signature) String() string {
	return strings.ToUpper(hex.EncodeToString(s))
}

// Validate returns an error if the value is invalid.
func (s Signature) Validate() error {
	if _, err := s.signature(); err != nil {
		return err
	}

	return nil
}

// signature converts a Signature to a schnorr.Signature.
func (s Signature) signature() (*schnorr.Signature, error) {
	signature, err := schnorr.ParseSignature(s)
	if err != nil {
		return nil, NewError(ErrParseError, err.Error())
	}

	return signature, nil
}

// R returns the R part of the signature.
func (s Signature) R() Point {
	if len(s) < 33 {
		return []byte{}
	}
	return Point(s[0:33])
}

// S returns the S part of the signature.
func (s Signature) S() Scalar {
	if len(s) < 65 {
		return []byte{}
	}
	return Scalar(s[33:65])
}

// ComplaintSignature represents a signature (a1, a2, z) stored as bytes.
// It uses schnorr.ComplaintSignature as a base implementation for serialization and parsing.
type ComplaintSignature []byte

// ComplaintSignatures represents a slice of ComplaintSignature values.
type ComplaintSignatures []ComplaintSignature

// NewComplaintSignature constructs a ComplaintSignature from bytes.
func NewComplaintSignature(bytes []byte) (ComplaintSignature, error) {
	comSignature := ComplaintSignature(bytes)
	if err := comSignature.Validate(); err != nil {
		return nil, NewError(err, "invalid")
	}

	return comSignature, nil
}

// NewComplaintSignatureFromComponents generates a complaint signature from 2 Points (A1, A2) and Scalar (Z).
// It returns a ComplaintSignature and an error, if any.
func NewComplaintSignatureFromComponents(rawA1 Point, rawA2 Point, rawZ Scalar) (ComplaintSignature, error) {
	a1, err := rawA1.jacobianPoint()
	if err != nil {
		return nil, NewError(err, "parse A1")
	}

	a2, err := rawA2.jacobianPoint()
	if err != nil {
		return nil, NewError(err, "parse A2")
	}

	z := rawZ.modNScalar()

	return NewComplaintSignatureFromType(schnorr.NewComplaintSignature(a1, a2, z)), nil
}

// NewComplaintSignatureFromType parses a schnorr.ComplaintSignature into a ComplaintSignature.
func NewComplaintSignatureFromType(signature *schnorr.ComplaintSignature) ComplaintSignature {
	return signature.Serialize()
}

// Marshal needed for protobuf compatibility
func (cs ComplaintSignature) Marshal() ([]byte, error) {
	return cs, nil
}

// Unmarshal needed for protobuf compatibility
func (cs *ComplaintSignature) Unmarshal(data []byte) error {
	*cs = data
	return nil
}

// MarshalJSON converts the ComplaintSignature to its JSON representation.
func (cs ComplaintSignature) MarshalJSON() ([]byte, error) {
	str := strings.ToUpper(hex.EncodeToString(cs))
	jbz := make([]byte, len(str)+2)
	jbz[0] = '"'
	copy(jbz[1:], str)
	jbz[len(jbz)-1] = '"'
	return jbz, nil
}

// UnmarshalJSON parses a JSON string into a ComplaintSignature.
func (cs *ComplaintSignature) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid hex string: %s", data)
	}
	bz2, err := hex.DecodeString(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}
	*cs = bz2
	return nil
}

// Bytes returns the underlying byte slice of the ComplaintSignature.
func (cs ComplaintSignature) Bytes() []byte {
	return cs
}

// String returns the hexadecimal representation of the ComplaintSignature in uppercase.
func (cs ComplaintSignature) String() string {
	return strings.ToUpper(hex.EncodeToString(cs))
}

// Validate returns an error if the value is invalid.
func (cs ComplaintSignature) Validate() error {
	if _, err := cs.complaintSignature(); err != nil {
		return err
	}

	return nil
}

// complaintSignature converts a ComplaintSignature to a schnorr.ComplaintSignature.
func (cs ComplaintSignature) complaintSignature() (*schnorr.ComplaintSignature, error) {
	// No need to check error as the caller should validate it first.
	signature, err := schnorr.ParseComplaintSignature(cs)
	if err != nil {
		return nil, NewError(ErrParseError, err.Error())
	}

	return signature, nil
}

// A1 returns the A1 part of the complaint signature.
func (cs ComplaintSignature) A1() Point {
	if len(cs) < 33 {
		return []byte{}
	}
	return Point(cs[0:33])
}

// A2 returns the A2 part of the complaint signature.
func (cs ComplaintSignature) A2() Point {
	if len(cs) < 66 {
		return []byte{}
	}
	return Point(cs[33:66])
}

// Z returns the Z part of the complaint signature.
func (cs ComplaintSignature) Z() Scalar {
	if len(cs) < 98 {
		return []byte{}
	}
	return Scalar(cs[66:98])
}

// KeyPair represents a key pair consisting of a private key and a public key.
type KeyPair struct {
	PrivKey Scalar
	PubKey  Point
}

// KeyPairs represents a slice of KeyPair values.
type KeyPairs []KeyPair

// CommitmentIDE represents a commitment issued by a participant in the signing process
//
// Fields:
// - ID: A NonZeroScalar identifier for the participant (uint64)
// - D: The hiding nonce commitment is represented by the type Point
// - E: The binding nonce commitment is represented by the type Point
type CommitmentIDE struct {
	ID MemberID
	D  Point
	E  Point
}

// CommitmentIDEList is a slice of CommitmentIDE structs.
//
// This type represents a list of commitments issued by each participant in the signing process.
// Each element in the list indicates an identifier and two commitment PublicKey values as a tuple
// <i, hiding_nonce_commitment, binding_nonce_commitment>.
//
// Please note that this list must be sorted in ascending order by identifier
type CommitmentIDEList []CommitmentIDE

// Len returns the number of elements in the CommitmentIDEList.
func (b CommitmentIDEList) Len() int {
	return len(b)
}

// Sort sorts the CommitmentIDEList in ascending order by the identifier (ID) of each CommitmentIDE.
// It also checks for repeated elements and returns an error if any are found.
func (b CommitmentIDEList) Sort() error {
	sort.Slice(b, func(i, j int) bool {
		return b[i].ID < b[j].ID
	})

	for i := 0; i < len(b)-1; i++ {
		if b[i].ID == b[i+1].ID {
			return fmt.Errorf("CommitmentIDEList: sorting fail because repeated element found at ID = %v", b[i].ID)
		}
	}

	return nil
}

// EncSecretShare represents a structure for storing an encrypted secret share.
// It contains the encrypted value `Value` and the corresponding nonce `Nonce`
// used in the Elgamal encryption process. The `Value` field holds the encrypted
// data, and `Nonce` is used to ensure the security and uniqueness of the encryption.
type EncSecretShare []byte

func NewEncSecretShare(value []byte, nonce []byte) (EncSecretShare, error) {
	enc := EncSecretShare(append(value, nonce...))
	if err := enc.Validate(); err != nil {
		return nil, err
	}

	return enc, nil
}

// Value return the value part of EncSecretShare
func (e EncSecretShare) Value() []byte {
	return e[0:32]
}

// Value return the nonce part of EncSecretShare
func (e EncSecretShare) Nonce() []byte {
	return e[32:48]
}

// Clone creates a deep copy of the EncSecretShare instance.
func (e EncSecretShare) Clone() EncSecretShare {
	bz := make([]byte, len(e))
	copy(bz, e)
	return bz
}

// Marshal needed for protobuf compatibility
func (e EncSecretShare) Marshal() ([]byte, error) {
	return e, nil
}

// Unmarshal needed for protobuf compatibility
func (e *EncSecretShare) Unmarshal(data []byte) error {
	*e = data
	return nil
}

// MarshalJSON converts the EncSecretShare to its JSON representation.
func (e EncSecretShare) MarshalJSON() ([]byte, error) {
	str := strings.ToUpper(hex.EncodeToString(e))
	jbz := make([]byte, len(str)+2)
	jbz[0] = '"'
	copy(jbz[1:], str)
	jbz[len(jbz)-1] = '"'
	return jbz, nil
}

// UnmarshalJSON parses a JSON string into a EncSecretShare.
func (e *EncSecretShare) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid hex string: %s", data)
	}
	bz2, err := hex.DecodeString(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}
	*e = bz2
	return nil
}

// Bytes returns the underlying byte slice of the EncSecretShare.
func (e EncSecretShare) Bytes() []byte {
	return e
}

// String returns the hexadecimal representation of the EncSecretShare in uppercase.
func (e EncSecretShare) String() string {
	return strings.ToUpper(hex.EncodeToString(e))
}

// Validate checks the integrity and validity of the EncSecretShare instance.
// It ensures that the encrypted value and nonce have the correct expected sizes.
func (e EncSecretShare) Validate() error {
	if len(e) != 48 {
		return fmt.Errorf("EncSecretShare: invalid size")
	}
	return nil
}

// EncSecretShares is a slice of EncSecretShare. It's used for storing multiple
// encrypted secret shares. This type is particularly useful when dealing with
// scenarios where multiple pieces of data need to be encrypted, such as in
// threshold cryptography or secure multiparty computations, where each participant
// might have their own encrypted share of a secret.
type EncSecretShares []EncSecretShare

// Clone creates a deep copy of the EncSecretShares slice.
// It iterates through the slice, cloning each EncSecretShare to ensure
// that modifications to the cloned slice do not affect the original EncSecretShares.
func (es EncSecretShares) Clone() EncSecretShares {
	copied := make([]EncSecretShare, len(es))
	for i, e := range es {
		copied[i] = e.Clone()
	}
	return copied
}

// Validate iterates through each EncSecretShare in the EncSecretShares slice,
// performing validation checks on each EncSecretShare.
func (es EncSecretShares) Validate() error {
	var err error
	for i, e := range es {
		err = e.Validate()
		if err != nil {
			return NewError(err, fmt.Sprintf("index %d error", i))
		}
	}
	return err
}
