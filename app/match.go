package band

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"

	"github.com/bandprotocol/chain/v3/app/mempool"
	bandtsskeeper "github.com/bandprotocol/chain/v3/x/bandtss/keeper"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

// feedsSubmitSignalPriceTxMatchHandler is a function that returns the match function for the Feeds SubmitSignalPriceTx.
func feedsSubmitSignalPriceTxMatchHandler(
	cdc codec.Codec,
	authzKeeper *authzkeeper.Keeper,
	feedsMsgServer feedstypes.MsgServer,
) mempool.TxMatchFn {
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
		return isSuccess(feedsMsgServer.SubmitSignalPrices(ctx, msg))
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
		return true
	default:
		return false
	}
}

// tssTxMatchHandler is a function that returns the match function for the TSS Tx.
func tssTxMatchHandler(
	cdc codec.Codec,
	authzKeeper *authzkeeper.Keeper,
	bandtssKeeper *bandtsskeeper.Keeper,
	tssMsgServer tsstypes.MsgServer,
) mempool.TxMatchFn {
	return func(ctx sdk.Context, tx sdk.Tx) bool {
		msgs := tx.GetMsgs()
		if len(msgs) == 0 {
			return false
		}
		for _, msg := range msgs {
			if !isValidTSSTxMsg(ctx, msg, cdc, authzKeeper, bandtssKeeper, tssMsgServer) {
				return false
			}
		}
		return true
	}
}

// isValidTSSTxMsg return true if the message is a valid for TSS Tx.
func isValidTSSTxMsg(
	ctx sdk.Context,
	msg sdk.Msg,
	cdc codec.Codec,
	authzKeeper *authzkeeper.Keeper,
	bandtssKeeper *bandtsskeeper.Keeper,
	tssMsgServer tsstypes.MsgServer,
) bool {
	switch msg := msg.(type) {
	case *tsstypes.MsgSubmitDKGRound1:
		return isSuccess(tssMsgServer.SubmitDKGRound1(ctx, msg))
	case *tsstypes.MsgSubmitDKGRound2:
		return isSuccess(tssMsgServer.SubmitDKGRound2(ctx, msg))
	case *tsstypes.MsgConfirm:
		return isSuccess(tssMsgServer.Confirm(ctx, msg))
	case *tsstypes.MsgComplain:
		return isSuccess(tssMsgServer.Complain(ctx, msg))
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

		return isSuccess(tssMsgServer.SubmitDEs(ctx, msg))
	case *tsstypes.MsgSubmitSignature:
		return isSuccess(tssMsgServer.SubmitSignature(ctx, msg))
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
			if !isValidTSSTxMsg(ctx, m, cdc, authzKeeper, bandtssKeeper, tssMsgServer) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// oracleReportTxMatchHandler is a function that returns the match function for the oracle report tx.
func oracleReportTxMatchHandler(
	cdc codec.Codec,
	authzKeeper *authzkeeper.Keeper,
	oracleMsgServer oracletypes.MsgServer,
) mempool.TxMatchFn {
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
		return isSuccess(oracleMsgServer.ReportData(ctx, msg))
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

		return true
	default:
		return false
	}
}

func isSuccess(_ any, err error) bool {
	return err == nil
}
