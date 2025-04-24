package band

import (
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"

	"github.com/bandprotocol/chain/v3/app/mempool"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

// DefaultLaneMatchHandler is a function that returns the match function for the default lane.
func DefaultLaneMatchHandler() mempool.TxMatchFn {
	return func(_ sdk.Context, _ sdk.Tx) bool {
		return true
	}
}

// CreateLanes creates the lanes for the Band mempool.
func CreateLanes(app *BandApp) []*mempool.Lane {
	// feedsLane handles feeds submit signal price transactions.
	// Each transaction has a gas limit of 2%, and the total gas limit for the lane is 50%.
	// It uses SenderNonceMempool to ensure transactions are ordered by sender and nonce, with no per-sender tx limit.
	feedsLane := mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		"feedsLane",
		mempool.NewTxMatchFn([]sdk.Msg{&feedstypes.MsgSubmitSignalPrices{}}, true),
		math.LegacyMustNewDecFromStr("0.02"),
		math.LegacyMustNewDecFromStr("0.5"),
		sdkmempool.NewSenderNonceMempool(sdkmempool.SenderNonceMaxTxOpt(0)),
		nil,
	)

	// tssLane handles TSS transactions.
	// Each transaction has a gas limit of 2%, and the total gas limit for the lane is 20%.
	tssLane := mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		"tssLane",
		mempool.NewTxMatchFn(
			[]sdk.Msg{
				&tsstypes.MsgSubmitDKGRound1{},
				&tsstypes.MsgSubmitDKGRound2{},
				&tsstypes.MsgConfirm{},
				&tsstypes.MsgComplain{},
				&tsstypes.MsgSubmitDEs{},
				&tsstypes.MsgSubmitSignature{},
			},
			true,
		),
		math.LegacyMustNewDecFromStr("0.02"),
		math.LegacyMustNewDecFromStr("0.2"),
		sdkmempool.DefaultPriorityMempool(),
		nil,
	)

	// oracleRequestLane handles oracle request data transactions.
	// Each transaction has a gas limit of 10%, and the total gas limit for the lane is 10%.
	// It is blocked if the oracle report lane exceeds its limit.
	oracleRequestLane := mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		"oracleRequestLane",
		mempool.NewTxMatchFn(
			[]sdk.Msg{
				&oracletypes.MsgRequestData{},
				&channeltypes.MsgRecvPacket{}, // TODO: Only match oracle request packet
			},
			false,
		),
		math.LegacyMustNewDecFromStr("0.1"),
		math.LegacyMustNewDecFromStr("0.1"),
		sdkmempool.DefaultPriorityMempool(),
		nil,
	)

	// oracleReportLane handles oracle report data transactions.
	// Each transaction has a gas limit of 5%, and the total gas limit for the lane is 20%.
	// It block the oracle request lane if it exceeds its limit.
	oracleReportLane := mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		"oracleReportLane",
		mempool.NewTxMatchFn([]sdk.Msg{&oracletypes.MsgReportData{}}, true),
		math.LegacyMustNewDecFromStr("0.05"),
		math.LegacyMustNewDecFromStr("0.2"),
		sdkmempool.DefaultPriorityMempool(),
		func(isLaneLimitExceeded bool) {
			oracleRequestLane.SetBlocked(isLaneLimitExceeded)
		},
	)

	// defaultLane handles all other transactions.
	// Each transaction has a gas limit of 10%, and the total gas limit for the lane is 10%.
	defaultLane := mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		"defaultLane",
		DefaultLaneMatchHandler(),
		math.LegacyMustNewDecFromStr("0.1"),
		math.LegacyMustNewDecFromStr("0.1"),
		sdkmempool.DefaultPriorityMempool(),
		nil,
	)

	return []*mempool.Lane{feedsLane, tssLane, oracleRequestLane, oracleReportLane, defaultLane}
}
