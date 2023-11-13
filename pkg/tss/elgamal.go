package tss

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

// Encrypt encrypts the given value using the key.
// encrypted value = Hash(key) + value
// It returns the encrypted value as a Scalar.
func Encrypt(value Scalar, key Point, nonces ...[]byte) (EncSecretShare, error) {
	var nonceBytes []byte
	var err error
	if len(nonces) == 0 || len(nonces[0]) != 16 {
		nonceBytes, err = RandomBytes(16)
		if err != nil {
			return EncSecretShare{}, err
		}
	} else {
		nonceBytes = nonces[0]
	}

	encValue, err := EncryptHKDF(value.Bytes(), Hash(key), nonceBytes)
	if err != nil {
		return EncSecretShare{}, err
	}

	return EncSecretShare{Value: encValue, Nonce: nonceBytes}, nil
}

// Decrypt decrypts the given encrypted value using the key.
// value = encrypted value - Hash(key)
// It returns the decrypted value as a Scalar.
func Decrypt(e EncSecretShare, key Point) (Scalar, error) {
	return DecryptHKDF(e, Hash(key))
}

func EncryptHKDF(shareBytes, aesKey, nonceBytes []byte) ([]byte, error) {
	if len(shareBytes) != 32 || len(aesKey) != 32 || len(nonceBytes) != 16 {
		return nil, fmt.Errorf("some input's size is invalid")
	}

	// Derive the AES key using HKDF with SHA512
	hkdfReader := hkdf.New(sha512.New, aesKey, nil, nil)
	finalAESKey := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, finalAESKey); err != nil {
		return nil, NewError(err, "failed to derive AES key")
	}

	// Create a new AES-CTR cipher object
	blockCipher, err := aes.NewCipher(finalAESKey)
	if err != nil {
		return nil, NewError(err, "failed to create AES cipher")
	}

	stream := cipher.NewCTR(blockCipher, nonceBytes)

	// Perform the encryption
	encrypted := make([]byte, len(shareBytes))
	stream.XORKeyStream(encrypted, shareBytes)

	return encrypted, nil
}

func DecryptHKDF(e EncSecretShare, aesKey []byte) ([]byte, error) {
	err := e.Validate()
	if err != nil {
		return nil, NewError(err, "at DecryptHKDF")
	}
	if len(aesKey) != 32 {
		return nil, errors.New("DecryptHKDF: aesKey's size is invalid")
	}

	// Derive the AES key using HKDF with SHA512
	hkdfReader := hkdf.New(sha512.New, aesKey, nil, nil)
	finalAESKey := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, finalAESKey); err != nil {
		return nil, NewError(err, "failed to derive AES key")
	}

	// Create a new AES-CTR cipher object
	blockCipher, err := aes.NewCipher(finalAESKey)
	if err != nil {
		return nil, NewError(err, "failed to create AES cipher")
	}
	stream := cipher.NewCTR(blockCipher, e.Nonce)

	// Perform the decryption
	decrypted := make([]byte, 32)
	stream.XORKeyStream(decrypted, e.Value)
	return decrypted, nil
}
