package tss_test

import (
	"math/rand"
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
)

func BenchmarkEncrypt(b *testing.B) {
	value := testutil.HexDecode("fc93f14f4e3e4e15378e2c65ba1986494a3f54b7c135dd21d67a44435332eb71")
	key := testutil.HexDecode("035db2a125a23300bef24e57883f547503ab2598a99ed07d65d482b4ea1ff8ed26")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tss.Encrypt(value, key)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	enc := tss.EncSecretShare{
		Value: testutil.HexDecode("cb0b29556849ad4219a5bb6fd7e12ac15805c9166371bcf2c4e931eeaf502807"),
		Nonce: testutil.HexDecode("d8e4136601557341913837f01885d307"),
	}
	key := testutil.HexDecode("64540a84e00ca07eb2f34bfa98caf96c8db3b09918427bca2863ee0b2d6df31f")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tss.Decrypt(enc, key)
	}
}

func BenchmarkHash(b *testing.B) {
	tests := []struct {
		name       string
		numOfBytes int
	}{{
		name:       "1 byte",
		numOfBytes: 1,
	}, {
		name:       "4 bytes",
		numOfBytes: 4,
	}, {
		name:       "32 bytes",
		numOfBytes: 32,
	}, {
		name:       "128 bytes",
		numOfBytes: 128,
	}, {
		name:       "1024 bytes",
		numOfBytes: 1024,
	}, {
		name:       "8096 bytes",
		numOfBytes: 8096,
	}}

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			rand.Seed(0)
			bytes := make([]byte, test.numOfBytes)
			rand.Read(bytes)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				tss.Hash(bytes)
			}
		})
	}
}
