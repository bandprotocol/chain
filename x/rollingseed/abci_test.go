package rollingseed_test

import (
	"encoding/hex"
	"testing"

	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func fromHex(hexStr string) []byte {
	res, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	return res
}

func TestRollingSeedCorrect(t *testing.T) {
	app, ctx, _ := testapp.CreateTestInput(false)

	// Initially rolling seed should be all zeros.
	require.Equal(
		t,
		fromHex("0000000000000000000000000000000000000000000000000000000000000000"),
		app.RollingseedKeeper.GetRollingSeed(ctx),
	)
	// Every begin block, the rolling seed should get updated.
	app.BeginBlocker(ctx, abci.RequestBeginBlock{
		Hash: fromHex("0100000000000000000000000000000000000000000000000000000000000000"),
	})
	require.Equal(
		t,
		fromHex("0000000000000000000000000000000000000000000000000000000000000001"),
		app.RollingseedKeeper.GetRollingSeed(ctx),
	)
	app.BeginBlocker(ctx, abci.RequestBeginBlock{
		Hash: fromHex("0200000000000000000000000000000000000000000000000000000000000000"),
	})
	require.Equal(
		t,
		fromHex("0000000000000000000000000000000000000000000000000000000000000102"),
		app.RollingseedKeeper.GetRollingSeed(ctx),
	)
	app.BeginBlocker(ctx, abci.RequestBeginBlock{
		Hash: fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
	})
	require.Equal(
		t,
		fromHex("00000000000000000000000000000000000000000000000000000000000102ff"),
		app.RollingseedKeeper.GetRollingSeed(ctx),
	)
}
