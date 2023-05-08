package types

// Coefficient is the type-safe unique identifier type for coefficient.
type Coefficient []byte

// Coefficients is the type-safe unique identifier type for coefficients.
type Coefficients []Coefficient

// Point is the type-safe unique identifier type for point.
type Point []byte

// Points is the type-safe unique identifier type for points.
type Points []Point

// PublicKey is the type-safe unique identifier type for public key.
type PublicKey []byte

// PublicKeys is the type-safe unique identifier type for public keys.
type PublicKeys []PublicKey

// PrivateKey is the type-safe unique identifier type for private key.
type PrivateKey []byte

// PrivateKeys is the type-safe unique identifier type for private keys.
type PrivateKeys []PrivateKey

// Signature is the type-safe unique identifier type for signature.
type Signature []byte

type KeyPair struct {
	PrivateKey PrivateKey
	PublicKey  PublicKey
}

type KeyPairs []KeyPair
