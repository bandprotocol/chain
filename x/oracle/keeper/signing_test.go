package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/testing/testapp"
	bandtsstypes "github.com/bandprotocol/chain/v2/x/bandtss/types"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

func TestGetSetSigningResult(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	rid := types.RequestID(123)
	signingResult := types.SigningResult{
		SigningID:      bandtsstypes.SigningID(456),
		ErrorCodespace: "",
		ErrorCode:      0,
	}

	// Set the signing result by request ID
	k.SetSigningResult(ctx, rid, signingResult)

	// Get the signing result associated with the request ID
	got, err := k.GetSigningResult(ctx, rid)
	require.NoError(t, err)
	require.Equal(t, signingResult, got)
}
