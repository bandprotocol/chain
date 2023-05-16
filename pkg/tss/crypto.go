package tss

import (
	"github.com/ethereum/go-ethereum/crypto"
)

// Encrypt encrypts the given value using the symmetric key.
// encrypted value = Hash(symmetric key) + value
// It returns the encrypted value as a Scalar.
func Encrypt(value Scalar, keySym PublicKey) Scalar {
	k := Scalar(Hash(keySym)).Parse()
	v := value.Parse()

	res := k.Add(v).Bytes()
	return res[:]
}

// Decrypt decrypts the given encrypted value using the symmetric key.
// value = encrypted value - Hash(symmetric key)
// It returns the decrypted value as a Scalar.
func Decrypt(encValue Scalar, keySym PublicKey) Scalar {
	k := Scalar(Hash(keySym)).Parse()
	ev := encValue.Parse()

	res := k.Negate().
		Add(ev).
		Bytes()

	return res[:]
}

// Hash calculates the Keccak-256 hash of the given data.
// It returns the hash value as a byte slice.
func Hash(data ...[]byte) []byte {
	return crypto.Keccak256(data...)
}
