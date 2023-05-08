package tss

import (
	"github.com/bandprotocol/chain/v2/x/tss/types"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/ethereum/go-ethereum/crypto"
)

func GenerateKeyPair() (types.KeyPair, error) {
	key, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return types.KeyPair{}, err
	}

	return types.KeyPair{
		PrivateKey: key.Serialize(),
		PublicKey:  key.PubKey().SerializeCompressed(),
	}, nil
}
func GenerateKeyPairs(n uint64) (types.KeyPairs, error) {
	var kps types.KeyPairs
	for i := uint64(0); i < n; i++ {
		kp, err := GenerateKeyPair()
		if err != nil {
			return nil, err
		}

		kps = append(kps, kp)
	}

	return kps, nil
}

func PublicKey(privKey types.PrivateKey) types.PublicKey {
	pk := secp256k1.PrivKeyFromBytes(privKey)
	return pk.PubKey().SerializeCompressed()
}

func Hash(salt []byte, data ...[]byte) []byte {
	return crypto.Keccak256(data...)
}

func Hash1(data ...[]byte) []byte {
	return Hash([]byte("round1A0"), data...)
}

func Hash2(data ...[]byte) []byte {
	return Hash([]byte("round1Sk"), data...)
}
