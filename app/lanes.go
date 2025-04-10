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
	feedskeeper "github.com/bandprotocol/chain/v3/x/feeds/keeper"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	oraclekeeper "github.com/bandprotocol/chain/v3/x/oracle/keeper"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
	tsskeeper "github.com/bandprotocol/chain/v3/x/tss/keeper"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

// FeedsLaneMatchHandler is a function that returns the match function for the Feeds lane.
func FeedsLaneMatchHandler(
	cdc codec.Codec,
	authzKeeper *authzkeeper.Keeper,
	feedsMsgServer feedstypes.MsgServer,
) func(sdk.Context, sdk.Tx) bool {
	return func(ctx sdk.Context, tx sdk.Tx) bool {
		msgs := tx.GetMsgs()
		if len(msgs) == 0 {
			return false
		}
		for _, msg := range msgs {
			if !isValidMsgSubmitSignalPrices(ctx, msg, cdc, authzKeeper, feedsMsgServer) {
				return false
			}
		}
		return true
	}
}

// isValidMsgSubmitSignalPrices return true if the message is a valid feeds' MsgSubmitSignalPrices.
func isValidMsgSubmitSignalPrices(
	ctx sdk.Context,
	msg sdk.Msg,
	cdc codec.Codec,
	authzKeeper *authzkeeper.Keeper,
	feedsMsgServer feedstypes.MsgServer,
) bool {
	switch msg := msg.(type) {
	case *feedstypes.MsgSubmitSignalPrices:
		if _, err := feedsMsgServer.SubmitSignalPrices(ctx, msg); err != nil {
			return false
		}
	case *authz.MsgExec:
		msgs, err := msg.GetMessages()
		if err != nil {
			return false
		}

		grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
		if err != nil {
			return false
		}

		for _, m := range msgs {
			signers, _, err := cdc.GetMsgV1Signers(m)
			if err != nil {
				return false
			}
			// Check if this grantee have authorization for the message.
			cap, _ := authzKeeper.GetAuthorization(
				ctx,
				grantee,
				sdk.AccAddress(signers[0]),
				sdk.MsgTypeURL(m),
			)
			if cap == nil {
				return false
			}

			// Check if this message should be free or not.
			if !isValidMsgSubmitSignalPrices(ctx, m, cdc, authzKeeper, feedsMsgServer) {
				return false
			}
		}
	default:
		return false
	}

	return true
}

// TssLaneMatchHandler is a function that returns the match function for the TSS lane.
func TssLaneMatchHandler(
	cdc codec.Codec,
	authzKeeper *authzkeeper.Keeper,
	bandtssKeeper *bandtsskeeper.Keeper,
	tssMsgServer tsstypes.MsgServer,
) func(sdk.Context, sdk.Tx) bool {
	return func(ctx sdk.Context, tx sdk.Tx) bool {
		msgs := tx.GetMsgs()
		if len(msgs) == 0 {
			return false
		}
		for _, msg := range msgs {
			if !isValidTssLaneMsg(ctx, msg, cdc, authzKeeper, bandtssKeeper, tssMsgServer) {
				return false
			}
		}
		return true
	}
}

// isValidTssLaneMsg return true if the message is a valid for TSS lane.
func isValidTssLaneMsg(
	ctx sdk.Context,
	msg sdk.Msg,
	cdc codec.Codec,
	authzKeeper *authzkeeper.Keeper,
	bandtssKeeper *bandtsskeeper.Keeper,
	tssMsgServer tsstypes.MsgServer,
) bool {
	switch msg := msg.(type) {
	case *tsstypes.MsgSubmitDKGRound1:
		if _, err := tssMsgServer.SubmitDKGRound1(ctx, msg); err != nil {
			return false
		}
	case *tsstypes.MsgSubmitDKGRound2:
		if _, err := tssMsgServer.SubmitDKGRound2(ctx, msg); err != nil {
			return false
		}
	case *tsstypes.MsgConfirm:
		if _, err := tssMsgServer.Confirm(ctx, msg); err != nil {
			return false
		}
	case *tsstypes.MsgComplain:
		if _, err := tssMsgServer.Complain(ctx, msg); err != nil {
			return false
		}
	case *tsstypes.MsgSubmitDEs:
		acc, err := sdk.AccAddressFromBech32(msg.Sender)
		if err != nil {
			return false
		}

		currentGroupID := bandtssKeeper.GetCurrentGroup(ctx).GroupID
		incomingGroupID := bandtssKeeper.GetIncomingGroupID(ctx)
		if !bandtssKeeper.HasMember(ctx, acc, currentGroupID) &&
			!bandtssKeeper.HasMember(ctx, acc, incomingGroupID) {
			return false
		}

		if _, err := tssMsgServer.SubmitDEs(ctx, msg); err != nil {
			return false
		}
	case *tsstypes.MsgSubmitSignature:
		if _, err := tssMsgServer.SubmitSignature(ctx, msg); err != nil {
			return false
		}
	case *authz.MsgExec:
		msgs, err := msg.GetMessages()
		if err != nil {
			return false
		}

		grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
		if err != nil {
			return false
		}

		for _, m := range msgs {
			signers, _, err := cdc.GetMsgV1Signers(m)
			if err != nil {
				return false
			}
			// Check if this grantee have authorization for the message.
			cap, _ := authzKeeper.GetAuthorization(
				ctx,
				grantee,
				sdk.AccAddress(signers[0]),
				sdk.MsgTypeURL(m),
			)
			if cap == nil {
				return false
			}

			// Check if this message should be free or not.
			if !isValidTssLaneMsg(ctx, m, cdc, authzKeeper, bandtssKeeper, tssMsgServer) {
				return false
			}
		}
	default:
		return false
	}

	return true
}

// oracleLaneMatchHandler is a function that returns the match function for the oracle lane.
func oracleLaneMatchHandler(
	cdc codec.Codec,
	authzKeeper *authzkeeper.Keeper,
	oracleMsgServer oracletypes.MsgServer,
) func(sdk.Context, sdk.Tx) bool {
	return func(ctx sdk.Context, tx sdk.Tx) bool {
		msgs := tx.GetMsgs()
		if len(msgs) == 0 {
			return false
		}
		for _, msg := range msgs {
			if !isValidMsgReportData(ctx, msg, cdc, authzKeeper, oracleMsgServer) {
				return false
			}
		}
		return true
	}
}

// isValidMsgReportData return true if the message is a valid oracle's MsgReportData.
func isValidMsgReportData(
	ctx sdk.Context,
	msg sdk.Msg,
	cdc codec.Codec,
	authzKeeper *authzkeeper.Keeper,
	oracleMsgServer oracletypes.MsgServer,
) bool {
	switch msg := msg.(type) {
	case *oracletypes.MsgReportData:
		if _, err := oracleMsgServer.ReportData(ctx, msg); err != nil {
			return false
		}
	case *authz.MsgExec:
		msgs, err := msg.GetMessages()
		if err != nil {
			return false
		}

		grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
		if err != nil {
			return false
		}

		for _, m := range msgs {
			signers, _, err := cdc.GetMsgV1Signers(m)
			if err != nil {
				return false
			}
			// Check if this grantee have authorization for the message.
			cap, _ := authzKeeper.GetAuthorization(
				ctx,
				grantee,
				sdk.AccAddress(signers[0]),
				sdk.MsgTypeURL(m),
			)
			if cap == nil {
				return false
			}

			// Check if this message should be free or not.
			if !isValidMsgReportData(ctx, m, cdc, authzKeeper, oracleMsgServer) {
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
func CreateLanes(app *BandApp) (feedsLane, tssLane, oracleLane, defaultLane *mempool.Lane) {
	// 1. Create the signer extractor. This is used to extract the expected signers from
	// a transaction. Each lane can have a different signer extractor if needed.
	signerAdapter := sdkmempool.NewDefaultSignerExtractionAdapter()

	feedsMsgServer := feedskeeper.NewMsgServerImpl(app.FeedsKeeper)
	tssMsgServer := tsskeeper.NewMsgServerImpl(app.TSSKeeper)
	oracleMsgServer := oraclekeeper.NewMsgServerImpl(app.OracleKeeper)

	feedsLane = mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		signerAdapter,
		"feedsLane",
		FeedsLaneMatchHandler(app.appCodec, &app.AuthzKeeper, feedsMsgServer),
		math.LegacyMustNewDecFromStr("0.05"),
		math.LegacyMustNewDecFromStr("0.3"),
		sdkmempool.NewSenderNonceMempool(sdkmempool.SenderNonceMaxTxOpt(10)),
	)

	tssLane = mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		signerAdapter,
		"tssLane",
		TssLaneMatchHandler(app.appCodec, &app.AuthzKeeper, &app.BandtssKeeper, tssMsgServer),
		math.LegacyMustNewDecFromStr("0.05"),
		math.LegacyMustNewDecFromStr("0.2"),
		sdkmempool.DefaultPriorityMempool(),
	)

	oracleLane = mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		signerAdapter,
		"oracleLane",
		oracleLaneMatchHandler(app.appCodec, &app.AuthzKeeper, oracleMsgServer),
		math.LegacyMustNewDecFromStr("0.05"),
		math.LegacyMustNewDecFromStr("0.2"),
		sdkmempool.DefaultPriorityMempool(),
	)

	defaultLane = mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		signerAdapter,
		"defaultLane",
		DefaultLaneMatchHandler(),
		math.LegacyMustNewDecFromStr("0.3"),
		math.LegacyMustNewDecFromStr("0.3"),
		sdkmempool.DefaultPriorityMempool(),
	)

	return
}
