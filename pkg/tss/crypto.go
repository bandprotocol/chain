package tss

import (
	"github.com/bandprotocol/chain/v2/x/tss/types"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/ethereum/go-ethereum/crypto"
)

func PublicKey(privKeyBytes types.PrivateKey) types.PublicKey {
	privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)
	return privKey.PubKey().SerializeCompressed()
}

func Encrypt(value []byte, key types.PrivateKey) []byte {
	hash := crypto.Keccak256(key)

	var k secp256k1.ModNScalar
	_ = k.SetByteSlice(hash)

	var v secp256k1.ModNScalar
	_ = v.SetByteSlice(value)

	res := k.Add(&v).Bytes()
	return res[:]
}

func Decrypt(encValue []byte, key types.PrivateKey) []byte {
	hash := crypto.Keccak256(key)

	var k secp256k1.ModNScalar
	_ = k.SetByteSlice(hash)

	var ev secp256k1.ModNScalar
	_ = ev.SetByteSlice(encValue)

	res := k.Negate().Add(&ev).Bytes()
	return res[:]
}
