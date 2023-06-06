package tss_test

import (
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

func BenchmarkEncrypt(b *testing.B) {
	value := hexDecode("fc93f14f4e3e4e15378e2c65ba1986494a3f54b7c135dd21d67a44435332eb71")
	key := hexDecode("035db2a125a23300bef24e57883f547503ab2598a99ed07d65d482b4ea1ff8ed26")

	for i := 0; i < b.N; i++ {
		tss.Encrypt(value, key)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	value := hexDecode("d47a459f272be3d22e54af5a0a45ea8318e88f2c3c767962b2b5f9ba53d9922d")
	key := hexDecode("035db2a125a23300bef24e57883f547503ab2598a99ed07d65d482b4ea1ff8ed26")

	for i := 0; i < b.N; i++ {
		tss.Decrypt(value, key)
	}
}

func BenchmarkHash(b *testing.B) {
	data := []byte("data")

	for i := 0; i < b.N; i++ {
		tss.Hash(data)
	}
}
