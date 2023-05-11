package tss_test

import (
	"encoding/hex"
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/stretchr/testify/assert"
)

func TestEncryptAndDecrypt(t *testing.T) {
	key := "4ac3ad151305074ba80e6a6abd44a5280a0502e9f06afd3e5aaad455c181ef57"
	k, err := hex.DecodeString(key)
	assert.NoError(t, err)

	expectedValue := "e463de6047df20228442e02c1ae58daf95e74e7a5763a94f8afe4d3b8bf97eb8"
	ev, err := hex.DecodeString(expectedValue)
	assert.NoError(t, err)

	ec := tss.Encrypt(ev, k)
	value := tss.Decrypt(ec, k)

	assert.Equal(t, ev, value)
}
