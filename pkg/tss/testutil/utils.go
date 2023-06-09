package testutil

import (
	"encoding/hex"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

func HexDecode(str string) []byte {
	b, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return b
}

func GetSlot(from tss.MemberID, to tss.MemberID) tss.MemberID {
	slot := to - 1
	if from < to {
		slot--
	}

	return slot
}

func PublicKey(privKey tss.PrivateKey) tss.PublicKey {
	pubKey, err := privKey.PublicKey()
	if err != nil {
		panic(err)
	}

	return pubKey
}

func CopySlice[T ~[]U, U ~[]byte](src T) T {
	dst := make(T, len(src))

	for i := range src {
		dst[i] = Copy(src[i])
	}

	return dst
}

func Copy[T ~[]byte](src T) T {
	dst := make(T, len(src))
	copy(dst, src)

	return dst
}
