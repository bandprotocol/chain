package crypto

import (
	"encoding/hex"
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/ethereum/go-ethereum/crypto"
)

func TryFunc() (result secp256k1.JacobianPoint) {
	k := hexToModNScalar("d74bf844b0862475103d96a611cf2d898447e288d34b360bc885cb8ce7c00575")
	secp256k1.ScalarBaseMultNonConst(k, &result)
	fmt.Println("aaaa", result.Z)

	result.ToAffine()

	fmt.Println("bbbb", result.Z)
	var a []byte
	a = append(a, result.X.Bytes()[:]...)
	a = append(a, result.Y.Bytes()[:]...)

	hash := crypto.Keccak256(a)

	fmt.Println(fmt.Sprintf("¥®¨£ˆƒ˙™ˆ¨˙ƒ™¢ˆ˙ƒ™ˆ¨  %x", hash))

	return
}

func hexToModNScalar(s string) *secp256k1.ModNScalar {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("invalid hex in source file: " + s)
	}
	var scalar secp256k1.ModNScalar
	if overflow := scalar.SetByteSlice(b); overflow {
		panic("hex in source file overflows mod N scalar: " + s)
	}
	return &scalar
}
