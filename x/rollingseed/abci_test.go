package rollingseed_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/core/header"

	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
)

func fromHex(hexStr string) []byte {
	res, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	return res
}

func TestRollingSeedCorrect(t *testing.T) {
	dir := sdktestutil.GetTempDir(t)
	app := bandtesting.SetupWithCustomHome(false, dir)
	ctx := app.BaseApp.NewUncachedContext(false, cmtproto.Header{ChainID: bandtesting.ChainID})
	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: app.LastBlockHeight() + 1})
	require.NoError(t, err)

	// Initially rolling seed should be all zeros.
	require.Equal(
		t,
		fromHex("0000000000000000000000000000000000000000000000000000000000000000"),
		app.RollingseedKeeper.GetRollingSeed(ctx),
	)
	// Every begin block, the rolling seed should get updated.
	_, err = app.BeginBlocker(ctx.WithHeaderInfo(
		header.Info{Hash: fromHex("0100000000000000000000000000000000000000000000000000000000000000")},
	))
	require.NoError(t, err)
	require.Equal(
		t,
		fromHex("0000000000000000000000000000000000000000000000000000000000000001"),
		app.RollingseedKeeper.GetRollingSeed(ctx),
	)

	_, err = app.BeginBlocker(ctx.WithHeaderInfo(
		header.Info{Hash: fromHex("0200000000000000000000000000000000000000000000000000000000000000")},
	))
	require.NoError(t, err)
	require.Equal(
		t,
		fromHex("0000000000000000000000000000000000000000000000000000000000000102"),
		app.RollingseedKeeper.GetRollingSeed(ctx),
	)

	_, err = app.BeginBlocker(ctx.WithHeaderInfo(
		header.Info{Hash: fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")},
	))
	require.NoError(t, err)
	require.Equal(
		t,
		fromHex("00000000000000000000000000000000000000000000000000000000000102ff"),
		app.RollingseedKeeper.GetRollingSeed(ctx),
	)
}
