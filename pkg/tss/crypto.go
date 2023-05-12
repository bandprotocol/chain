package tss

import (
	"github.com/ethereum/go-ethereum/crypto"
)

func Encrypt(value Scalar, keySym PublicKey) Scalar {
	k := Scalar(Hash(keySym)).Parse()
	v := value.Parse()

	res := k.Add(v).Bytes()
	return res[:]
}

func Decrypt(encValue Scalar, keySym PublicKey) Scalar {
	k := Scalar(Hash(keySym)).Parse()
	ev := encValue.Parse()

	res := k.Negate().Add(ev).Bytes()
	return res[:]
}

func Hash(data ...[]byte) []byte {
	return crypto.Keccak256(data...)
}
