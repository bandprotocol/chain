package emitter

import (
	abci "github.com/cometbft/cometbft/abci/types"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v3/hooks/common"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
	restaketypes "github.com/bandprotocol/chain/v3/x/restake/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
	tunneltypes "github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func parseEvents(events []abci.Event) common.EvMap {
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
func (h *Hook) handleMsg(ctx sdk.Context, txHash []byte, msg sdk.Msg, events []abci.Event, detail common.JsDict) {
	evMap := parseEvents(events)
	switch msg := msg.(type) {
	case *oracletypes.MsgRequestData:
		h.handleMsgRequestData(ctx, txHash, msg, evMap, detail)
	case *oracletypes.MsgReportData:
		h.handleMsgReportData(ctx, txHash, msg, "")
	case *oracletypes.MsgCreateDataSource:
		h.handleMsgCreateDataSource(ctx, txHash, evMap, detail)
	case *oracletypes.MsgCreateOracleScript:
		h.handleMsgCreateOracleScript(ctx, txHash, evMap, detail)
	case *oracletypes.MsgEditDataSource:
		h.handleMsgEditDataSource(ctx, txHash, msg)
	case *oracletypes.MsgEditOracleScript:
		h.handleMsgEditOracleScript(ctx, txHash, msg)
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
	case *govv1beta1.MsgSubmitProposal:
		h.handleV1beta1MsgSubmitProposal(ctx, txHash, msg, evMap, detail)
	case *govv1.MsgSubmitProposal:
		h.handleMsgSubmitProposal(ctx, txHash, msg, evMap, detail)
	case *govv1beta1.MsgVote:
		h.handleV1beta1MsgVote(ctx, txHash, msg, detail)
	case *govv1.MsgVote:
		h.handleMsgVote(ctx, txHash, msg, detail)
	case *govv1beta1.MsgVoteWeighted:
		h.handleV1beta1MsgVoteWeighted(ctx, txHash, msg, detail)
	case *govv1.MsgVoteWeighted:
		h.handleMsgVoteWeighted(ctx, txHash, msg, detail)
	case *govv1beta1.MsgDeposit:
		h.handleV1beta1MsgDeposit(ctx, txHash, msg, detail)
	case *govv1.MsgDeposit:
		h.handleMsgDeposit(ctx, txHash, msg, detail)
	case *channeltypes.MsgRecvPacket:
		h.handleMsgRecvPacket(ctx, txHash, msg, events, evMap, detail)
	case *transfertypes.MsgTransfer:
		h.handleMsgTransfer(ctx, txHash, msg, evMap, detail)
	case *clienttypes.MsgCreateClient:
		h.handleMsgCreatClient(msg, detail)
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
	case *channeltypes.MsgTimeout:
		h.handleMsgTimeout(ctx, msg)
	case *authz.MsgGrant:
		h.handleMsgGrant(msg, detail)
	case *authz.MsgRevoke:
		h.handleMsgRevoke(msg, detail)
	case *authz.MsgExec:
		h.handleMsgExec(ctx, txHash, msg, events, detail)
	case *bandtsstypes.MsgActivate:
		h.handleBandtssMsgActivate(ctx, msg)
	case *bandtsstypes.MsgRequestSignature:
		h.handleTSSEventCreateSigning(ctx, evMap)
		h.handleTSSEventRequestSignature(ctx, evMap)
		h.handleBandtssEventSigningRequestCreated(ctx, evMap)
	case *feedstypes.MsgVote:
		h.handleFeedsMsgVote(ctx, msg, evMap)
	case *feedstypes.MsgSubmitSignalPrices:
		h.handleFeedsMsgSubmitSignalPrices(ctx, txHash, msg, "")
	case *feedstypes.MsgUpdateReferenceSourceConfig:
		h.handleFeedsMsgUpdateReferenceSourceConfig(ctx, msg)
	case *tsstypes.MsgSubmitSignature:
		h.handleTSSEventSubmitSignature(ctx, evMap)
	case *tunneltypes.MsgCreateTunnel:
		h.handleTunnelMsgCreateTunnel(ctx, txHash, msg, evMap)
	case *tunneltypes.MsgUpdateSignalsAndInterval:
		h.handleTunnelMsgUpdateSignalsAndInterval(ctx, evMap)
	case *tunneltypes.MsgDepositToTunnel:
		h.handleTunnelMsgDepositToTunnel(ctx, txHash, msg)
	case *tunneltypes.MsgWithdrawFromTunnel:
		h.handleTunnelMsgWithdrawFromTunnel(ctx, txHash, msg)
	case *tunneltypes.MsgTriggerTunnel:
		tunnelSenderFeesMap := getTunnelSenderFeesMap(ctx, *h, events)
		h.handleTunnelMsgTriggerTunnel(ctx, msg, evMap, tunnelSenderFeesMap)
	default:
		break
	}

	for _, event := range events {
		h.handleMsgEvent(ctx, txHash, event)
	}
}

func (h *Hook) handleMsgEvent(ctx sdk.Context, txHash []byte, event abci.Event) {
	evMap := parseEvents([]abci.Event{event})
	switch event.Type {
	case banktypes.EventTypeTransfer:
		h.handleMsgEventTypeTransfer(evMap)
	case restaketypes.EventTypeCreateVault:
		h.handleRestakeEventCreateVault(ctx, evMap)
	case restaketypes.EventTypeLockPower:
		h.handleRestakeEventLockPower(ctx, txHash, evMap)
	case restaketypes.EventTypeStake:
		h.handleRestakeEventStake(ctx, evMap)
	case restaketypes.EventTypeUnstake:
		h.handleRestakeEventUnstake(ctx, evMap)
	case restaketypes.EventTypeDeactivateVault:
		h.handleRestakeEventDeactivateVault(ctx, evMap)
	case tunneltypes.EventTypeActivateTunnel:
		h.handleTunnelEventTypeActivateTunnel(ctx, evMap)
	case tunneltypes.EventTypeDeactivateTunnel:
		h.handleTunnelEventTypeDeactivateTunnel(ctx, evMap)
	case tsstypes.EventTypeSetMemberIsActive:
		h.handleTSSSetMember(ctx, evMap)
	}
}

func (h *Hook) handleBeginBlockEndBlockEvent(
	ctx sdk.Context,
	event abci.Event,
	eventIdx int,
	eventQuerier *EventQuerier,
	tunnelSenderFeesMap map[string]Fees,
) {
	evMap := parseEvents([]abci.Event{event})
	switch event.Type {
	case bandtsstypes.EventTypeInactiveStatus:
		h.handleBandtssEventInactiveStatus(ctx, evMap)
	case bandtsstypes.EventTypeGroupTransition:
		h.handleBandtssEventGroupTransition(ctx, eventIdx, eventQuerier)
	case bandtsstypes.EventTypeGroupTransitionSuccess:
		h.handleBandtssEventGroupTransitionSuccess(ctx, evMap)
	case bandtsstypes.EventTypeGroupTransitionFailed:
		h.handleBandtssEventGroupTransitionFailed(ctx, evMap)
	case bandtsstypes.EventTypeMemberDeleted:
		h.handleBandtssEventDeleteMember(ctx, evMap)
	case bandtsstypes.EventTypeSigningRequestCreated:
		h.handleBandtssEventSigningRequestCreated(ctx, evMap)
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
	case tsstypes.EventTypeCreateSigning:
		h.handleTSSEventCreateSigning(ctx, evMap)
	case tsstypes.EventTypeRequestSignature:
		h.handleTSSEventRequestSignature(ctx, evMap)
	case tsstypes.EventTypeSigningSuccess:
		h.handleTSSEventSigningSuccess(ctx, evMap)
	case tsstypes.EventTypeSigningFailed:
		h.handleTSSEventSigningFailed(ctx, evMap)
	case tsstypes.EventTypeCreateGroup,
		tsstypes.EventTypeRound2Success,
		tsstypes.EventTypeRound3Success,
		tsstypes.EventTypeExpiredGroup,
		tsstypes.EventTypeComplainSuccess,
		tsstypes.EventTypeRound3Failed:
		groupIDs := evMap[event.Type+"."+tsstypes.AttributeKeyGroupID]
		for _, gid := range groupIDs {
			h.handleTSSSetGroup(ctx, tss.GroupID(common.Atoi(gid)))
		}
	case tsstypes.EventTypeSetMemberIsActive:
		h.handleTSSSetMember(ctx, evMap)
	case tunneltypes.EventTypeProducePacketSuccess:
		h.handleTunnelEventTypeProducePacketSuccess(ctx, evMap, tunnelSenderFeesMap)
	case tunneltypes.EventTypeActivateTunnel:
		h.handleTunnelEventTypeActivateTunnel(ctx, evMap)
	case tunneltypes.EventTypeDeactivateTunnel:
		h.handleTunnelEventTypeDeactivateTunnel(ctx, evMap)
	default:
		break
	}
}
