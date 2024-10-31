package types_test

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func TestEncodeSigning(t *testing.T) {
	ctx := sdk.Context{}.
		WithBlockHeader(cmtproto.Header{Time: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)}).
		WithChainID("")

	got := types.EncodeSigning(ctx, 1, []byte("originator"), []byte("message"))
	strHex := hex.EncodeToString(got)
	expected := "" +
		"c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470" +
		"bac0e8e27c59b287045fc0a3df1b9bc08bca23b9c7d4e8d21f6c311f67a7ef4b" +
		"000000005e0be100" +
		"0000000000000001" +
		"6d657373616765"

	require.Equal(t, expected, strHex)
}
