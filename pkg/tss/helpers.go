package tss

import (
	"github.com/bandprotocol/chain/v2/pkg/tss/internal/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// ConcatBytes concatenates multiple byte slices into a single byte slice.
func ConcatBytes(data ...[]byte) []byte {
	var res []byte
	for _, b := range data {
		res = append(res, b...)
	}

	return res
}

// GenerateKeyPairs generates a specified number of key pairs.
// It returns a slice of KeyPair values and an error, if any.
func GenerateKeyPairs(n uint64) (KeyPairs, error) {
	var kps KeyPairs
	for i := uint64(0); i < n; i++ {
		kp, err := GenerateKeyPair()
		if err != nil {
			return nil, NewError(err, "number: %d", i)
		}

		kps = append(kps, kp)
	}

	return kps, nil
}

// GenerateKeyPair generates a new key pair.
// It returns a KeyPair value and an error, if any.
func GenerateKeyPair() (KeyPair, error) {
	key, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return KeyPair{}, NewError(ErrGenerateKeyPairFailed, err.Error())
	}

	return KeyPair{
		PrivKey: key.Serialize(),
		PubKey:  key.PubKey().SerializeCompressed(),
	}, nil
}

// GenerateNonce generates a nonce value using the provided private key, hash, and iterator.
// It returns the nonce value as a Scalar.
func GenerateNonce(privKey PrivateKey, hash []byte) (Scalar, Point) {
	nonce := generateNonce(privKey, hash, 0)

	var pubNonce secp256k1.JacobianPoint
	secp256k1.ScalarBaseMultNonConst(nonce, &pubNonce)

	return ParseScalar(nonce), ParsePoint(&pubNonce)
}

// generateNonce generates a nonce value using the provided private key, hash, and iterator.
// It returns the nonce value as a *secp256k1.ModNScalar.
func generateNonce(privKey PrivateKey, hash []byte, iterator uint32) *secp256k1.ModNScalar {
	return secp256k1.NonceRFC6979(
		privKey,
		hash,
		schnorr.RFC6979ExtraDataV0[:],
		nil,
		iterator,
	)
}
