package bandtss_test

import (
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	bandtesting "github.com/bandprotocol/chain/v2/testing"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestReplaceGroups(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, true)
	tssKeeper, bandtssKeeper := app.TSSKeeper, app.BandtssKeeper

	// Set new block time
	ctx = ctx.WithBlockTime(time.Now().UTC())

	now := time.Now().UTC()
	beforenow := now.Add(time.Duration(-5) * time.Minute)

	signingID := tss.SigningID(1)
	currentGroupID := tss.GroupID(1)
	incomingGroupID := tss.GroupID(2)

	// Set up initial state for testing
	currentGroup := tsstypes.Group{
		ID:            currentGroupID,
		Size_:         5,
		Threshold:     3,
		PubKey:        testutil.HexDecode("0260aa1c85288f77aeaba5d02e984d987b16dd7f6722544574a03d175b48d8b83b"),
		Status:        tsstypes.GROUP_STATUS_ACTIVE,
		CreatedHeight: 1,
	}
	incomingGroup := tsstypes.Group{
		ID:            incomingGroupID,
		Size_:         7,
		Threshold:     4,
		PubKey:        testutil.HexDecode("02a37461c1621d12f2c436b98ffe95d6ff0fedc102e8b5b35a08c96b889cb448fd"),
		Status:        tsstypes.GROUP_STATUS_ACTIVE,
		CreatedHeight: 2,
	}

	signing := tsstypes.Signing{
		ID:     signingID,
		Status: tsstypes.SIGNING_STATUS_SUCCESS,
	}

	tssKeeper.SetGroup(ctx, currentGroup)
	tssKeeper.SetGroup(ctx, incomingGroup)
	tssKeeper.SetMember(ctx, tsstypes.Member{
		ID:      tss.MemberID(1),
		GroupID: currentGroupID,
		Address: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
	})
	tssKeeper.SetMember(ctx, tsstypes.Member{
		ID:      tss.MemberID(1),
		GroupID: incomingGroupID,
		Address: "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
	})

	bandtssKeeper.SetCurrentGroupID(ctx, currentGroupID)
	bandtssKeeper.SetGroupTransition(ctx, types.GroupTransition{
		SigningID:           signingID,
		CurrentGroupID:      currentGroupID,
		CurrentGroupPubKey:  currentGroup.PubKey,
		IncomingGroupID:     incomingGroupID,
		IncomingGroupPubKey: incomingGroup.PubKey,
		ExecTime:            beforenow,
		Status:              types.TRANSITION_STATUS_WAITING_EXECUTION,
	})
	tssKeeper.SetSigning(ctx, signing)

	// Call end block
	app.EndBlocker(ctx, abci.RequestEndBlock{Height: app.LastBlockHeight() + 1})

	_, found := bandtssKeeper.GetGroupTransition(ctx)
	require.False(t, found)
	require.Equal(t, incomingGroupID, bandtssKeeper.GetCurrentGroupID(ctx))
}
