package emitter

import (
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	proto "github.com/cosmos/gogoproto/proto"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/bandprotocol/chain/v2/hooks/common"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	bandtsstypes "github.com/bandprotocol/chain/v2/x/bandtss/types"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
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
		h.handleMsgRecvPacket(ctx, txHash, msg, evMap, log, detail)
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
	case *authz.MsgGrant:
		h.handleMsgGrant(msg, detail)
	case *authz.MsgRevoke:
		h.handleMsgRevoke(msg, detail)
	case *authz.MsgExec:
		h.handleMsgExec(ctx, txHash, msg, log, detail)
	case *bandtsstypes.MsgActivate:
		h.handleBandtssMsgActivate(ctx, msg)
	case *bandtsstypes.MsgHealthCheck:
		h.handleBandtssMsgHealthCheck(ctx, msg)
	case *bandtsstypes.MsgRequestSignature:
		h.handleEventRequestSignature(ctx, evMap)
	case *tsstypes.MsgSubmitDEs:
		h.handleTSSMsgSubmitDEs(ctx, msg)
	case *group.MsgCreateGroup:
		h.handleGroupMsgCreateGroup(ctx, evMap)
	case *group.MsgCreateGroupPolicy:
		h.handleGroupMsgCreateGroupPolicy(ctx, evMap)
	case *group.MsgCreateGroupWithPolicy:
		h.handleGroupMsgCreateGroupWithPolicy(ctx, evMap)
	case *group.MsgExec:
		h.handleGroupEventExec(ctx, evMap)
	case *group.MsgLeaveGroup:
		h.handleGroupMsgLeaveGroup(ctx, evMap)
	case *group.MsgSubmitProposal:
		h.handleGroupMsgSubmitProposal(ctx, evMap)
	case *group.MsgUpdateGroupAdmin:
		h.handleGroupMsgUpdateGroupAdmin(ctx, evMap)
	case *group.MsgUpdateGroupMembers:
		h.handleGroupMsgUpdateGroupMembers(ctx, msg, evMap)
	case *group.MsgUpdateGroupMetadata:
		h.handleGroupMsgUpdateGroupMetadata(ctx, evMap)
	case *group.MsgUpdateGroupPolicyAdmin:
		h.handleGroupMsgUpdateGroupPolicyAdmin(ctx, evMap)
	case *group.MsgUpdateGroupPolicyDecisionPolicy:
		h.handleGroupMsgUpdateGroupPolicyDecisionPolicy(ctx, evMap)
	case *group.MsgUpdateGroupPolicyMetadata:
		h.handleGroupMsgUpdateGroupPolicyMetadata(ctx, evMap)
	case *group.MsgVote:
		h.handleGroupMsgVote(ctx, msg, evMap)
		h.handleGroupEventExec(ctx, evMap)
	case *group.MsgWithdrawProposal:
		h.handleGroupMsgWithdrawProposal(ctx, evMap)
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
	case tsstypes.EventTypeRequestSignature:
		h.handleEventRequestSignature(ctx, evMap)
	case tsstypes.EventTypeSigningSuccess:
		h.handleEventSigningSuccess(ctx, evMap)
	case tsstypes.EventTypeSigningFailed:
		h.handleEventSigningFailed(ctx, evMap)
	case tsstypes.EventTypeExpiredSigning:
		h.handleEventExpiredSigning(ctx, evMap)
	case bandtsstypes.EventTypeInactiveStatus:
		address := sdk.MustAccAddressFromBech32(
			evMap[bandtsstypes.EventTypeInactiveStatus+"."+tsstypes.AttributeKeyAddress][0],
		)
		h.handleUpdateBandtssStatus(ctx, address)
	case tsstypes.EventTypeCreateGroup,
		tsstypes.EventTypeRound2Success,
		tsstypes.EventTypeRound3Success,
		tsstypes.EventTypeComplainSuccess,
		tsstypes.EventTypeComplainFailed,
		tsstypes.EventTypeExpiredGroup:

		gid := tss.GroupID(common.Atoi(evMap[event.Type+"."+tsstypes.AttributeKeyGroupID][0]))
		h.handleSetTSSGroup(ctx, gid)
	case bandtsstypes.EventTypeReplacement:
		if evMap[bandtsstypes.EventTypeReplacement+"."+bandtsstypes.AttributeKeyReplacementStatus][0] == "1" {
			h.handleInitBandtssReplacement(ctx)
		} else {
			// TODO: check EventTypeNewGroupActivate
			h.handleUpdateBandtssReplacementStatus(ctx)
		}
	case proto.MessageName(&group.EventProposalPruned{}):
		h.handleGroupEventProposalPruned(ctx, evMap)
	default:
		break
	}
}

func splitKeyWithTime(key []byte) (proposalID uint64, endTime time.Time) {
	lenTime := len(sdk.FormatTimeBytes(time.Now()))
	kv.AssertKeyLength(key[2:], 8+lenTime)

	endTime, err := sdk.ParseTimeBytes(key[2 : 2+lenTime])
	if err != nil {
		panic(err)
	}

	proposalID = sdk.BigEndianToUint64(key[2+lenTime:])
	return
}
