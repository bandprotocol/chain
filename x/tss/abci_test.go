package tss_test

import (
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	bandtesting "github.com/bandprotocol/chain/v2/testing"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestReplaceGroups(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, true)
	k := app.TSSKeeper

	// Set new block time
	ctx = ctx.WithBlockTime(time.Now().UTC())

	now := time.Now().UTC()
	beforenow := now.Add(time.Duration(-5) * time.Minute)

	signingID := tss.SigningID(1)
	currentGroupID := tss.GroupID(1)
	newGroupID := tss.GroupID(2)

	// Set up initial state for testing
	initialCurrentGroup := types.Group{
		ID:            currentGroupID,
		Size_:         5,
		Threshold:     3,
		PubKey:        testutil.HexDecode("0260aa1c85288f77aeaba5d02e984d987b16dd7f6722544574a03d175b48d8b83b"),
		Status:        types.GROUP_STATUS_ACTIVE,
		CreatedHeight: 1,
	}
	initialNewGroup := types.Group{
		ID:            newGroupID,
		Size_:         7,
		Threshold:     4,
		PubKey:        testutil.HexDecode("02a37461c1621d12f2c436b98ffe95d6ff0fedc102e8b5b35a08c96b889cb448fd"),
		Status:        types.GROUP_STATUS_ACTIVE,
		CreatedHeight: 2,
	}

	initialSigning := types.Signing{
		ID:     signingID,
		Status: types.SIGNING_STATUS_SUCCESS,
		// ... other fields ...
	}
	k.SetGroup(ctx, initialCurrentGroup)
	k.SetGroup(ctx, initialNewGroup)
	k.SetSigning(ctx, initialSigning)

	// Create a pending replace group with an execution time set 5 minutes before
	pendingReplaceGroup := types.Replacement{
		SigningID:      signingID,
		CurrentGroupID: currentGroupID,
		CurrentPubKey:  initialCurrentGroup.PubKey,
		NewGroupID:     newGroupID,
		NewPubKey:      initialNewGroup.PubKey,
		ExecTime:       now.Add(time.Duration(-5) * time.Minute),
	}

	nextID := k.GetNextReplacementCount(ctx)
	pendingReplaceGroup.ID = nextID
	k.SetReplacement(ctx, pendingReplaceGroup)

	k.InsertReplacementQueue(ctx, nextID, beforenow)

	// Call end block
	app.EndBlocker(ctx, abci.RequestEndBlock{Height: app.LastBlockHeight() + 1})

	got := k.MustGetGroup(ctx, currentGroupID)
	require.Equal(t, initialNewGroup.PubKey, got.PubKey)
}
