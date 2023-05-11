package tss_test

import (
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

func TestAAAAA(t *testing.T) {
	data, err := tss.GenerateRound1Data(1, 1, 3, []byte("11156101179830 29 171 230 26 197 169 217 200 154 35 128 180 121 192 141 183 43 132 103 139 196 154 251 139 50 176"))

	t.Log(err)
	t.Logf("%+v", data)
}
