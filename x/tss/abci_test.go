package tss_test

import (
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestReplaceGroups(t *testing.T) {
	app, ctx, _ := testapp.CreateTestInput(true)
	k := app.TSSKeeper

	// Set new block time
	ctx = ctx.WithBlockTime(time.Now())

	signingID := tss.SigningID(1)
	fromGroupID := tss.GroupID(1)
	toGroupID := tss.GroupID(2)

	// Set up initial state for testing
	initialFromGroup := types.Group{
		GroupID:       fromGroupID,
		Size_:         7,
		Threshold:     4,
		PubKey:        testutil.HexDecode("02a37461c1621d12f2c436b98ffe95d6ff0fedc102e8b5b35a08c96b889cb448fd"),
		Status:        types.GROUP_STATUS_ACTIVE,
		Fee:           sdk.NewCoins(sdk.NewInt64Coin("uband", 15)),
		CreatedHeight: 2,
	}
	initialToGroup := types.Group{
		GroupID:       toGroupID,
		Size_:         5,
		Threshold:     3,
		PubKey:        testutil.HexDecode("0260aa1c85288f77aeaba5d02e984d987b16dd7f6722544574a03d175b48d8b83b"),
		Status:        types.GROUP_STATUS_ACTIVE,
		Fee:           sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
		CreatedHeight: 1,
	}
	initialSigning := types.Signing{
		SigningID: signingID,
		Status:    types.SIGNING_STATUS_SUCCESS,
		// ... other fields ...
	}
	k.SetGroup(ctx, initialFromGroup)
	k.SetGroup(ctx, initialToGroup)
	k.SetSigning(ctx, initialSigning)

	// Create a pending replace group with an execution time set 5 minutes before
	pendingReplaceGroup1 := types.PendingReplaceGroup{
		SigningID:   signingID,
		FromGroupID: fromGroupID,
		ToGroupID:   toGroupID,
		ExecTime:    time.Now().UTC().Add(time.Duration(-5) * time.Minute),
	}
	// Create a pending replace group with an execution time set 5 minutes after
	pendingReplaceGroup2 := types.PendingReplaceGroup{
		SigningID:   signingID,
		FromGroupID: fromGroupID,
		ToGroupID:   toGroupID,
		ExecTime:    time.Now().UTC().Add(time.Duration(5) * time.Minute),
	}

	k.AddPendingReplaceGroup(ctx, pendingReplaceGroup1)
	k.AddPendingReplaceGroup(ctx, pendingReplaceGroup2)

	// Call end block
	app.EndBlocker(ctx, abci.RequestEndBlock{Height: app.LastBlockHeight() + 1})

	got := k.GetPendingReplaceGroups(ctx)
	require.Len(t, got, 1)
}
