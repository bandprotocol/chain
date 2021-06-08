package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	transfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/bandprotocol/chain/hooks/common"
	oracletypes "github.com/bandprotocol/chain/x/oracle/types"
)

func parseEvents(events sdk.StringEvents) common.EvMap {
	evMap := make(common.EvMap)
	for _, event := range events {
		for _, kv := range event.Attributes {
			key := event.Type + "." + kv.Key
			evMap[key] = append(evMap[key], kv.Value)
		}
	}
	return evMap
}

// handleMsg handles the given message by publishing relevant events and populates accounts
// that need balance update in 'h.accs'. Also fills in extra info for this message.
func (h *Hook) handleMsg(ctx sdk.Context, txHash []byte, msg sdk.Msg, log sdk.ABCIMessageLog, msgJson common.JsDict) {
	evMap := parseEvents(log.Events)
	switch msg := msg.(type) {
	case *oracletypes.MsgRequestData:
		h.handleMsgRequestData(ctx, txHash, msg, evMap, msgJson)
	case *oracletypes.MsgReportData:
		h.handleMsgReportData(ctx, txHash, msg, evMap)
	case *oracletypes.MsgCreateDataSource:
		h.handleMsgCreateDataSource(ctx, txHash, evMap, msgJson)
	case *oracletypes.MsgCreateOracleScript:
		h.handleMsgCreateOracleScript(ctx, txHash, evMap, msgJson)
	case *oracletypes.MsgEditDataSource:
		h.handleMsgEditDataSource(ctx, txHash, msg)
	case *oracletypes.MsgEditOracleScript:
		h.handleMsgEditOracleScript(ctx, txHash, msg)
	case *oracletypes.MsgAddReporter:
		h.handleMsgAddReporter(ctx, msg, msgJson)
	case *oracletypes.MsgRemoveReporter:
		h.handleMsgRemoveReporter(ctx, msg, msgJson)
	case *oracletypes.MsgActivate:
		h.handleMsgActivate(ctx, msg)
	case *stakingtypes.MsgCreateValidator:
		h.handleMsgCreateValidator(ctx, msg, msgJson)
	case *stakingtypes.MsgEditValidator:
		h.handleMsgEditValidator(ctx, msg, msgJson)
	case *stakingtypes.MsgDelegate:
		h.handleMsgDelegate(ctx, msg, msgJson)
	case *stakingtypes.MsgUndelegate:
		h.handleMsgUndelegate(ctx, msg, evMap, msgJson)
	case *stakingtypes.MsgBeginRedelegate:
		h.handleMsgBeginRedelegate(ctx, msg, evMap, msgJson)
	case *banktypes.MsgSend:
		h.handleMsgSend(msg)
	case *banktypes.MsgMultiSend:
		h.handleMsgMultiSend(msg)
	case *distrtypes.MsgWithdrawDelegatorReward:
		h.handleMsgWithdrawDelegatorReward(ctx, msg, evMap, msgJson)
	case *distrtypes.MsgSetWithdrawAddress:
		h.handleMsgSetWithdrawAddress(msg, msgJson)
	case *distrtypes.MsgWithdrawValidatorCommission:
		h.handleMsgWithdrawValidatorCommission(ctx, msg, evMap, msgJson)
	case *slashingtypes.MsgUnjail:
		h.handleMsgUnjail(ctx, msg)
	case *govtypes.MsgSubmitProposal:
		h.handleMsgSubmitProposal(ctx, txHash, msg, evMap, msgJson)
	case *govtypes.MsgVote:
		h.handleMsgVote(ctx, txHash, msg, msgJson)
	case *govtypes.MsgDeposit:
		h.handleMsgDeposit(ctx, txHash, msg, msgJson)
	case *channeltypes.MsgRecvPacket:
		h.handleMsgRecvPacket(ctx, txHash, msg, evMap, msgJson)
	case *transfertypes.MsgTransfer:
		h.handleMsgTransfer(ctx, msg, evMap)
	}
}

func (h *Hook) handleBeginBlockEndBlockEvent(ctx sdk.Context, event abci.Event) {
	events := sdk.StringifyEvents([]abci.Event{event})
	evMap := parseEvents(events)
	switch event.Type {
	case oracletypes.EventTypeResolve:
		h.handleEventRequestExecute(ctx, evMap)
	case slashingtypes.EventTypeSlash:
		h.handleEventSlash(ctx, evMap)
	case oracletypes.EventTypeDeactivate:
		h.handleEventDeactivate(ctx, evMap)
	case stakingtypes.EventTypeCompleteUnbonding:
		h.handleEventTypeCompleteUnbonding(ctx, evMap)
	case stakingtypes.EventTypeCompleteRedelegation:
		h.handEventTypeCompleteRedelegation(ctx)
	case govtypes.EventTypeInactiveProposal:
		h.handleEventInactiveProposal(evMap)
	case govtypes.EventTypeActiveProposal:
		h.handleEventTypeActiveProposal(ctx, evMap)
	case banktypes.EventTypeTransfer:
		h.handleEventTypeTransfer(evMap)
	case channeltypes.EventTypeSendPacket:
		h.handleEventSendPacket(ctx, evMap)
	default:
		break
	}
}

func (h *Hook) decodeMsgJson(ctx sdk.Context, msg sdk.Msg, msgJson common.JsDict) {
	switch msg := msg.(type) {
	case *oracletypes.MsgRequestData:
		decodeMsgRequestData(msg, msgJson)
	case *oracletypes.MsgReportData:
		decodeMsgReportData(msg, msgJson)
	case *oracletypes.MsgCreateDataSource:
		decodeMsgCreateDataSource(msg, msgJson)
	case *oracletypes.MsgCreateOracleScript:
		decodeMsgCreateOracleScript(msg, msgJson)
	case *oracletypes.MsgEditDataSource:
		decodeMsgEditDataSource(msg, msgJson)
	case *oracletypes.MsgEditOracleScript:
		decodeMsgEditOracleScript(msg, msgJson)
	case *oracletypes.MsgAddReporter:
		decodeMsgAddReporter(msg, msgJson)
	case *oracletypes.MsgRemoveReporter:
		decodeMsgRemoveReporter(msg, msgJson)
	case *oracletypes.MsgActivate:
		decodeMsgActivate(msg, msgJson)
	default:
		break
	}
}
