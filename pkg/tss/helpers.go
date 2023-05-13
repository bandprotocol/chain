package tss

import (
	"github.com/bandprotocol/chain/v2/pkg/tss/internal/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func ConcatBytes(data ...[]byte) []byte {
	var res []byte
	for _, b := range data {
		res = append(res, b...)
	}

	return res
}

func GenerateKeyPairs(n uint64) (KeyPairs, error) {
	var kps KeyPairs
	for i := uint64(0); i < n; i++ {
		kp, err := GenerateKeyPair()
		if err != nil {
			return nil, err
		}

		kps = append(kps, kp)
	}

	return kps, nil
}

func GenerateKeyPair() (KeyPair, error) {
	key, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return KeyPair{}, err
	}

	return KeyPair{
		PrivateKey: key.Serialize(),
		PublicKey:  key.PubKey().SerializeCompressed(),
	}, nil
}

func GenerateNonce(privKey PrivateKey, hash []byte, iterator uint32) Scalar {
	return ParseScalar(generateNonce(privKey, hash, iterator))
}

func generateNonce(privKey PrivateKey, hash []byte, iterator uint32) *secp256k1.ModNScalar {
	return secp256k1.NonceRFC6979(
		privKey,
		hash,
		schnorr.RFC6979ExtraDataV0[:],
		nil,
		iterator,
	)
}
