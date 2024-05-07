package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	bandtesting "github.com/bandprotocol/chain/v2/testing"
	bandtsstypes "github.com/bandprotocol/chain/v2/x/bandtss/types"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

func TestGetSetSigningResult(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, true)
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
