package tss_test

import (
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/stretchr/testify/assert"
)

func TestConcatBytes(t *testing.T) {
	res := tss.ConcatBytes([]byte("abc"), []byte("de"), []byte("fghi"))
	assert.Equal(t, []byte("abcdefghi"), res)
}

func TestGenerateKeyPair(t *testing.T) {
	_, err := tss.GenerateKeyPair()
	assert.Nil(t, err)
}
