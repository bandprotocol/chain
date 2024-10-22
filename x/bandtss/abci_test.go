package bandtss_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	band "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/pkg/tss/testutil"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

func init() {
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(sdk.GetConfig())
}

func TestReplaceGroups(t *testing.T) {
	dir := sdktestutil.GetTempDir(t)
	app := bandtesting.SetupWithCustomHome(false, dir)
	ctx := app.BaseApp.NewUncachedContext(false, cmtproto.Header{ChainID: bandtesting.ChainID})
	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: app.LastBlockHeight() + 1})
	require.NoError(t, err)

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
	_, err = app.EndBlocker(ctx.WithBlockHeight(ctx.BlockHeight() + 1))
	require.NoError(t, err)

	_, found := bandtssKeeper.GetGroupTransition(ctx)
	require.False(t, found)
	require.Equal(t, incomingGroupID, bandtssKeeper.GetCurrentGroupID(ctx))
}
