package crypto_test

import (
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/crypto"
)

func TestTryFunc(t *testing.T) {
	result := crypto.TryFunc()

	t.Log(result)
}
