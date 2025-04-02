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

// FeedsLaneMatchHandler is a function that returns the match function for the Feeds lane.
func FeedsLaneMatchHandler(
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
			if !isMsgSubmitSignalPrices(ctx, m, cdc, authzKeeper) {
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
	case *tsstypes.MsgSubmitDKGRound1:
		return true
	case *tsstypes.MsgSubmitDKGRound2:
		return true
	case *tsstypes.MsgConfirm:
		return true
	case *tsstypes.MsgComplain:
		return true
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
		return true
	case *tsstypes.MsgSubmitSignature:
		return true
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
			if !isTssLaneMsg(ctx, m, cdc, authzKeeper, bandtssKeeper) {
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
			if !isMsgReportData(ctx, m, cdc, authzKeeper) {
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

	feedsLane = mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		signerAdapter,
		"feedsLane",
		FeedsLaneMatchHandler(app.appCodec, &app.AuthzKeeper),
		math.LegacyMustNewDecFromStr("0.05"),
		math.LegacyMustNewDecFromStr("0.3"),
		sdkmempool.DefaultPriorityMempool(),
	)

	tssLane = mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		signerAdapter,
		"tssLane",
		TssLaneMatchHandler(app.appCodec, &app.AuthzKeeper, &app.BandtssKeeper),
		math.LegacyMustNewDecFromStr("0.05"),
		math.LegacyMustNewDecFromStr("0.2"),
		sdkmempool.DefaultPriorityMempool(),
	)

	oracleLane = mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		signerAdapter,
		"oracleLane",
		oracleLaneMatchHandler(app.appCodec, &app.AuthzKeeper),
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
