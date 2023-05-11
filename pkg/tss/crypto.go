package tss

import (
	"github.com/ethereum/go-ethereum/crypto"
)

func Encrypt(value Scalar, keySym PublicKey) Scalar {
	k := Scalar(crypto.Keccak256(keySym)).Parse()
	v := value.Parse()

	res := k.Add(v).Bytes()
	return res[:]
}

func Decrypt(encValue Scalar, keySym PublicKey) Scalar {
	k := Scalar(crypto.Keccak256(keySym)).Parse()
	ev := encValue.Parse()

	res := k.Negate().Add(ev).Bytes()
	return res[:]
}
