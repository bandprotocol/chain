package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

func TestGetSetSigningResult(t *testing.T) {
	dir := sdktestutil.GetTempDir(t)
	app := bandtesting.SetupWithCustomHome(false, dir)
	ctx := app.BaseApp.NewUncachedContext(false, cmtproto.Header{ChainID: bandtesting.ChainID})

	rid := types.RequestID(123)
	signingResult := types.SigningResult{
		SigningID:      bandtsstypes.SigningID(456),
		ErrorCodespace: "",
		ErrorCode:      0,
	}

	// Set the signing result by request ID
	app.OracleKeeper.SetSigningResult(ctx, rid, signingResult)

	// Get the signing result associated with the request ID
	got, err := app.OracleKeeper.GetSigningResult(ctx, rid)
	require.NoError(t, err)
	require.Equal(t, signingResult, got)
}
