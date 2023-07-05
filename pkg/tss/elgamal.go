package tss

import "github.com/decred/dcrd/dcrec/secp256k1/v4"

// Encrypt encrypts the given value using the key.
// encrypted value = Hash(key) + value
// It returns the encrypted value as a Scalar.
func Encrypt(value Scalar, key Point) (Scalar, error) {
	var k secp256k1.ModNScalar
	k.SetByteSlice(Hash(key))

	v := value.modNScalar()

	res := k.Add(v).Bytes()
	return res[:], nil
}

// Decrypt decrypts the given encrypted value using the key.
// value = encrypted value - Hash(key)
// It returns the decrypted value as a Scalar.
func Decrypt(encValue Scalar, key Point) (Scalar, error) {
	var k secp256k1.ModNScalar
	k.SetByteSlice(Hash(key))

	ev := encValue.modNScalar()
	res := k.Negate().Add(ev).Bytes()

	return res[:], nil
}
