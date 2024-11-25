package tss_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tsslib "github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss"
)

func TestEncoderPrefix(t *testing.T) {
	require.Equal(t, []byte(tss.TextMsgPrefix), tsslib.Hash([]byte("Text"))[:4])
}
