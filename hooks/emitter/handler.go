package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	transfertypes "github.com/cosmos/ibc-go/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/bandprotocol/chain/v2/hooks/common"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
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
func (h *Hook) handleMsg(ctx sdk.Context, txHash []byte, msg sdk.Msg, log sdk.ABCIMessageLog, detail common.JsDict) {
	evMap := parseEvents(log.Events)
	switch msg := msg.(type) {
	case *oracletypes.MsgRequestData:
		h.handleMsgRequestData(ctx, txHash, msg, evMap, detail)
	case *oracletypes.MsgReportData:
		h.handleMsgReportData(ctx, txHash, msg, evMap)
	case *oracletypes.MsgCreateDataSource:
		h.handleMsgCreateDataSource(ctx, txHash, evMap, detail)
	case *oracletypes.MsgCreateOracleScript:
		h.handleMsgCreateOracleScript(ctx, txHash, evMap, detail)
	case *oracletypes.MsgEditDataSource:
		h.handleMsgEditDataSource(ctx, txHash, msg)
	case *oracletypes.MsgEditOracleScript:
		h.handleMsgEditOracleScript(ctx, txHash, msg)
	case *oracletypes.MsgAddReporter:
		h.handleMsgAddReporter(ctx, msg, detail)
	case *oracletypes.MsgRemoveReporter:
		h.handleMsgRemoveReporter(ctx, msg, detail)
	case *oracletypes.MsgActivate:
		h.handleMsgActivate(ctx, msg)
	case *stakingtypes.MsgCreateValidator:
		h.handleMsgCreateValidator(ctx, msg, detail)
	case *stakingtypes.MsgEditValidator:
		h.handleMsgEditValidator(ctx, msg, detail)
	case *stakingtypes.MsgDelegate:
		h.handleMsgDelegate(ctx, msg, detail)
	case *stakingtypes.MsgUndelegate:
		h.handleMsgUndelegate(ctx, msg, evMap, detail)
	case *stakingtypes.MsgBeginRedelegate:
		h.handleMsgBeginRedelegate(ctx, msg, evMap, detail)
	case *banktypes.MsgSend:
		h.handleMsgSend(msg)
	case *banktypes.MsgMultiSend:
		h.handleMsgMultiSend(msg)
	case *distrtypes.MsgWithdrawDelegatorReward:
		h.handleMsgWithdrawDelegatorReward(ctx, msg, evMap, detail)
	case *distrtypes.MsgSetWithdrawAddress:
		h.handleMsgSetWithdrawAddress(msg, detail)
	case *distrtypes.MsgWithdrawValidatorCommission:
		h.handleMsgWithdrawValidatorCommission(ctx, msg, evMap, detail)
	case *slashingtypes.MsgUnjail:
		h.handleMsgUnjail(ctx, msg)
	case *govtypes.MsgSubmitProposal:
		h.handleMsgSubmitProposal(ctx, txHash, msg, evMap, detail)
	case *govtypes.MsgVote:
		h.handleMsgVote(ctx, txHash, msg, detail)
	case *govtypes.MsgDeposit:
		h.handleMsgDeposit(ctx, txHash, msg, detail)
	case *channeltypes.MsgRecvPacket:
		h.handleMsgRecvPacket(ctx, txHash, msg, evMap, detail)
	case *transfertypes.MsgTransfer:
		h.handleMsgTransfer(ctx, txHash, msg, evMap, detail)
	case *clienttypes.MsgCreateClient:
		h.handleMsgCreatClient(ctx, msg, detail)
	case *connectiontypes.MsgConnectionOpenConfirm:
		h.handleMsgConnectionOpenConfirm(ctx, msg)
	case *connectiontypes.MsgConnectionOpenAck:
		h.handleMsgConnectionOpenAck(ctx, msg)
	case *channeltypes.MsgChannelOpenInit:
		h.handleMsgChannelOpenInit(ctx, msg, evMap)
	case *channeltypes.MsgChannelOpenTry:
		h.handleMsgChannelOpenTry(ctx, msg, evMap)
	case *channeltypes.MsgChannelOpenAck:
		h.handleMsgChannelOpenAck(ctx, msg)
	case *channeltypes.MsgChannelOpenConfirm:
		h.handleMsgChannelOpenConfirm(ctx, msg)
	case *channeltypes.MsgChannelCloseInit:
		h.handleMsgChannelCloseInit(ctx, msg)
	case *channeltypes.MsgChannelCloseConfirm:
		h.handleMsgChannelCloseConfirm(ctx, msg)
	case *channeltypes.MsgAcknowledgement:
		h.handleMsgAcknowledgement(ctx, msg, evMap)
	default:
		break
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
