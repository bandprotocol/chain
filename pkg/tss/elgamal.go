package tss

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

// Encrypt takes a scalar value and a point key and returns an encrypted secret share.
// It uses a Nonce16Generator to generate a nonce for the encryption process.
func Encrypt(value Scalar, key Point, n16g Nonce16Generator) (EncSecretShare, error) {
	// Generate a 16-byte nonce using the provided Nonce16Generator
	nonceBytes, err := n16g.RandBytes16()
	if err != nil {
		return EncSecretShare{}, err
	}

	// Encrypt the value using HKDF-based encryption
	encValue, err := EncryptHKDF(value.Bytes(), Hash(key), nonceBytes)
	if err != nil {
		return EncSecretShare{}, err
	}

	return NewEncSecretShare(encValue, nonceBytes)
}

// Decrypt takes an encrypted secret share and a point key, and returns the decrypted scalar value.
// It delegates the decryption process to the DecryptHKDF function.
func Decrypt(e EncSecretShare, key Point) (Scalar, error) {
	// Call DecryptHKDF with the encrypted secret share and a hashed version of the key
	// Hash(key) generates an appropriate AES key from the Point key
	return DecryptHKDF(e, Hash(key))
}

// EncryptHKDF encrypts the given shareBytes using an AES key derived from the provided aesKey
// and nonceBytes using the HKDF algorithm.
func EncryptHKDF(shareBytes, aesKey, nonceBytes []byte) ([]byte, error) {
	if len(shareBytes) != 32 {
		return nil, fmt.Errorf("EncryptHKDF: share must be bytes32 but got bytes%d", len(shareBytes))
	}

	if len(aesKey) != 32 {
		return nil, fmt.Errorf("EncryptHKDF: aesKey must be bytes32 but got bytes%d", len(aesKey))
	}

	if len(nonceBytes) != 16 {
		return nil, fmt.Errorf("EncryptHKDF: nonce must be bytes16 but got bytes%d", len(nonceBytes))
	}

	// Derive the AES key using HKDF with SHA512
	hkdfReader := hkdf.New(sha512.New, aesKey, nil, nil)
	finalAESKey := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, finalAESKey); err != nil {
		return nil, NewError(err, "EncryptHKDF: failed to derive AES key")
	}

	// Create a new AES-CTR cipher object
	blockCipher, err := aes.NewCipher(finalAESKey)
	if err != nil {
		return nil, NewError(err, "EncryptHKDF: failed to create AES cipher")
	}

	stream := cipher.NewCTR(blockCipher, nonceBytes)

	// Perform the encryption
	encrypted := make([]byte, len(shareBytes))
	stream.XORKeyStream(encrypted, shareBytes)

	return encrypted, nil
}

// DecryptHKDF decrypts the given encrypted secret share using an AES key derived from the provided aesKey.
// It uses the HKDF algorithm for key derivation and AES in CTR mode for decryption.
func DecryptHKDF(e EncSecretShare, aesKey []byte) ([]byte, error) {
	err := e.Validate()
	if err != nil {
		return nil, NewError(err, "DecryptHKDF")
	}

	// Check if the provided AES key has the correct size
	if len(aesKey) != 32 {
		return nil, fmt.Errorf("DecryptHKDF: aesKey must be bytes32 but got bytes%d", len(aesKey))
	}

	// Derive the AES key using HKDF with SHA512
	hkdfReader := hkdf.New(sha512.New, aesKey, nil, nil)
	finalAESKey := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, finalAESKey); err != nil {
		return nil, NewError(err, "DecryptHKDF: failed to derive AES key")
	}

	// Create a new AES-CTR cipher object
	blockCipher, err := aes.NewCipher(finalAESKey)
	if err != nil {
		return nil, NewError(err, "DecryptHKDF: failed to create AES cipher")
	}
	stream := cipher.NewCTR(blockCipher, e.Nonce())

	// Perform the decryption
	decrypted := make([]byte, 32)
	stream.XORKeyStream(decrypted, e.Value())
	return decrypted, nil
}
