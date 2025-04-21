package band

import (
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"

	"github.com/bandprotocol/chain/v3/app/mempool"
	bandtsskeeper "github.com/bandprotocol/chain/v3/x/bandtss/keeper"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

// feedsLaneMatchHandler is a function that returns the match function for the Feeds lane.
func feedsLaneMatchHandler(
	cdc codec.Codec,
	authzKeeper *authzkeeper.Keeper,
) func(sdk.Context, sdk.Tx) bool {
	return func(ctx sdk.Context, tx sdk.Tx) bool {
		// Feeds lane only matches fee-less transactions
		gasTx, ok := tx.(sdk.FeeTx)
		if !ok {
			return false
		}

		if !gasTx.GetFee().IsZero() {
			return false
		}

		msgs := tx.GetMsgs()
		if len(msgs) == 0 {
			return false
		}
		for _, msg := range msgs {
			if !isMsgSubmitSignalPrices(ctx, msg, cdc, authzKeeper) {
				return false
			}
		}
		return true
	}
}

// isMsgSubmitSignalPrices return true if the message is a valid feeds' MsgSubmitSignalPrices.
func isMsgSubmitSignalPrices(
	ctx sdk.Context,
	msg sdk.Msg,
	cdc codec.Codec,
	authzKeeper *authzkeeper.Keeper,
) bool {
	switch msg := msg.(type) {
	case *feedstypes.MsgSubmitSignalPrices:
		return true
	case *authz.MsgExec:
		msgs, err := msg.GetMessages()
		if err != nil {
			return false
		}

		for _, m := range msgs {
			if !isMsgSubmitSignalPrices(ctx, m, cdc, authzKeeper) {
				return false
			}
		}
	default:
		return false
	}

	return true
}

// tssLaneMatchHandler is a function that returns the match function for the TSS lane.
func tssLaneMatchHandler(
	cdc codec.Codec,
	authzKeeper *authzkeeper.Keeper,
	bandtssKeeper *bandtsskeeper.Keeper,
) func(sdk.Context, sdk.Tx) bool {
	return func(ctx sdk.Context, tx sdk.Tx) bool {
		// TSS lane only matches fee-less transactions
		gasTx, ok := tx.(sdk.FeeTx)
		if !ok {
			return false
		}

		if !gasTx.GetFee().IsZero() {
			return false
		}

		msgs := tx.GetMsgs()
		if len(msgs) == 0 {
			return false
		}
		for _, msg := range msgs {
			if !isTssLaneMsg(ctx, msg, cdc, authzKeeper, bandtssKeeper) {
				return false
			}
		}
		return true
	}
}

// isTssLaneMsg return true if the message is a valid for TSS lane.
func isTssLaneMsg(
	ctx sdk.Context,
	msg sdk.Msg,
	cdc codec.Codec,
	authzKeeper *authzkeeper.Keeper,
	bandtssKeeper *bandtsskeeper.Keeper,
) bool {
	switch msg := msg.(type) {
	case *tsstypes.MsgSubmitDKGRound1,
		*tsstypes.MsgSubmitDKGRound2,
		*tsstypes.MsgConfirm,
		*tsstypes.MsgComplain,
		*tsstypes.MsgSubmitDEs,
		*tsstypes.MsgSubmitSignature:
		return true
	case *authz.MsgExec:
		msgs, err := msg.GetMessages()
		if err != nil {
			return false
		}

		for _, m := range msgs {
			if !isTssLaneMsg(ctx, m, cdc, authzKeeper, bandtssKeeper) {
				return false
			}
		}
	default:
		return false
	}

	return true
}

// oracleReportLaneMatchHandler is a function that returns the match function for the oracle lane.
func oracleReportLaneMatchHandler(
	cdc codec.Codec,
	authzKeeper *authzkeeper.Keeper,
) func(sdk.Context, sdk.Tx) bool {
	return func(ctx sdk.Context, tx sdk.Tx) bool {
		// Oracle lane only matches fee-less transactions
		gasTx, ok := tx.(sdk.FeeTx)
		if !ok {
			return false
		}

		if !gasTx.GetFee().IsZero() {
			return false
		}

		msgs := tx.GetMsgs()
		if len(msgs) == 0 {
			return false
		}
		for _, msg := range msgs {
			if !isMsgReportData(ctx, msg, cdc, authzKeeper) {
				return false
			}
		}
		return true
	}
}

// isMsgReportData return true if the message is a valid oracle's MsgReportData.
func isMsgReportData(
	ctx sdk.Context,
	msg sdk.Msg,
	cdc codec.Codec,
	authzKeeper *authzkeeper.Keeper,
) bool {
	switch msg := msg.(type) {
	case *oracletypes.MsgReportData:
		return true
	case *authz.MsgExec:
		msgs, err := msg.GetMessages()
		if err != nil {
			return false
		}

		for _, m := range msgs {
			if !isMsgReportData(ctx, m, cdc, authzKeeper) {
				return false
			}
		}
	default:
		return false
	}

	return true
}

// oracleRequestLaneMatchHandler is a function that returns the match function for the oracle request lane.
func oracleRequestLaneMatchHandler(
	cdc codec.Codec,
	authzKeeper *authzkeeper.Keeper,
) func(sdk.Context, sdk.Tx) bool {
	return func(ctx sdk.Context, tx sdk.Tx) bool {
		msgs := tx.GetMsgs()
		if len(msgs) == 0 {
			return false
		}
		for _, msg := range msgs {
			if !isMsgRequestData(ctx, msg, cdc, authzKeeper) {
				return false
			}
		}
		return true
	}
}

// isMsgRequestData return true if the message is a valid oracle's MsgRequestData.
func isMsgRequestData(
	ctx sdk.Context,
	msg sdk.Msg,
	cdc codec.Codec,
	authzKeeper *authzkeeper.Keeper,
) bool {
	switch msg := msg.(type) {
	case *oracletypes.MsgRequestData:
		return true
	case *authz.MsgExec:
		msgs, err := msg.GetMessages()
		if err != nil {
			return false
		}

		for _, m := range msgs {
			if !isMsgRequestData(ctx, m, cdc, authzKeeper) {
				return false
			}
		}
	default:
		return false
	}

	return true
}

// DefaultLaneMatchHandler is a function that returns the match function for the default lane.
func DefaultLaneMatchHandler() func(sdk.Context, sdk.Tx) bool {
	return func(_ sdk.Context, _ sdk.Tx) bool {
		return true
	}
}

// CreateLanes creates the lanes for the Band mempool.
func CreateLanes(app *BandApp) (feedsLane, tssLane, oracleReportLane, oracleRequestLane, defaultLane *mempool.Lane) {
	// Create the signer extractor. This is used to extract the expected signers from
	// a transaction. Each lane can have a different signer extractor if needed.
	signerExtractor := sdkmempool.NewDefaultSignerExtractionAdapter()

	// feedsLane handles feeds submit signal price transactions.
	// Each transaction has a gas limit of 2%, and the total gas limit for the lane is 50%.
	// It uses SenderNonceMempool to ensure transactions are ordered by sender and nonce, with no per-sender tx limit.
	feedsLane = mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		signerExtractor,
		"feedsLane",
		feedsLaneMatchHandler(app.appCodec, &app.AuthzKeeper),
		math.LegacyMustNewDecFromStr("0.02"),
		math.LegacyMustNewDecFromStr("0.5"),
		sdkmempool.NewSenderNonceMempool(sdkmempool.SenderNonceMaxTxOpt(0)),
		nil,
	)

	// tssLane handles TSS transactions.
	// Each transaction has a gas limit of 2%, and the total gas limit for the lane is 20%.
	tssLane = mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		signerExtractor,
		"tssLane",
		tssLaneMatchHandler(app.appCodec, &app.AuthzKeeper, &app.BandtssKeeper),
		math.LegacyMustNewDecFromStr("0.02"),
		math.LegacyMustNewDecFromStr("0.2"),
		sdkmempool.DefaultPriorityMempool(),
		nil,
	)

	// oracleRequestLane handles oracle request data transactions.
	// Each transaction has a gas limit of 10%, and the total gas limit for the lane is 10%.
	// It is blocked if the oracle report lane exceeds its limit.
	oracleRequestLane = mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		signerExtractor,
		"oracleRequestLane",
		oracleRequestLaneMatchHandler(app.appCodec, &app.AuthzKeeper),
		math.LegacyMustNewDecFromStr("0.1"),
		math.LegacyMustNewDecFromStr("0.1"),
		sdkmempool.DefaultPriorityMempool(),
		nil,
	)

	// oracleReportLane handles oracle report data transactions.
	// Each transaction has a gas limit of 5%, and the total gas limit for the lane is 20%.
	// It block the oracle request lane if it exceeds its limit.
	oracleReportLane = mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		signerExtractor,
		"oracleReportLane",
		oracleReportLaneMatchHandler(app.appCodec, &app.AuthzKeeper),
		math.LegacyMustNewDecFromStr("0.05"),
		math.LegacyMustNewDecFromStr("0.2"),
		sdkmempool.DefaultPriorityMempool(),
		func(isLaneLimitExceeded bool) {
			oracleRequestLane.SetBlocked(isLaneLimitExceeded)
		},
	)

	// defaultLane handles all other transactions.
	// Each transaction has a gas limit of 10%, and the total gas limit for the lane is 10%.
	defaultLane = mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		signerExtractor,
		"defaultLane",
		DefaultLaneMatchHandler(),
		math.LegacyMustNewDecFromStr("0.1"),
		math.LegacyMustNewDecFromStr("0.1"),
		sdkmempool.DefaultPriorityMempool(),
		nil,
	)

	return
}
