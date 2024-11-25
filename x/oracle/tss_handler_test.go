package oracle_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/oracle"
)

func TestEncoderPrefix(t *testing.T) {
	require.Equal(t, []byte(oracle.EncoderProtoPrefix), tss.Hash([]byte("Proto"))[:4])
	require.Equal(t, []byte(oracle.EncoderFullABIPrefix), tss.Hash([]byte("FullABI"))[:4])
	require.Equal(t, []byte(oracle.EncoderPartialABIPrefix), tss.Hash([]byte("PartialABI"))[:4])
}
