package testutil

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

type MockNonce16Generator struct {
	MockGenerateFunc func() ([]byte, error)
}

func (m MockNonce16Generator) RandBytes16() ([]byte, error) {
	if m.MockGenerateFunc != nil {
		return m.MockGenerateFunc()
	}
	// Default behavior if generateFunc is not set
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, tss.NewError(tss.ErrRandomError, "read bytes")
	}
	return b, nil
}

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

func Point(scalar tss.Scalar) tss.Point {
	point := scalar.Point()
	return point
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
