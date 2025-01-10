package emitter

import (
	"encoding/hex"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	feegranttypes "cosmossdk.io/x/feegrant"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v3/hooks/common"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
	restaketypes "github.com/bandprotocol/chain/v3/x/restake/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
	tunneltypes "github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func DecodeMsg(msg sdk.Msg, detail common.JsDict) {
	switch msg := msg.(type) {
	case *oracletypes.MsgRequestData:
		DecodeMsgRequestData(msg, detail)
	case *oracletypes.MsgReportData:
		DecodeMsgReportData(msg, detail)
	case *oracletypes.MsgCreateDataSource:
		DecodeMsgCreateDataSource(msg, detail)
	case *oracletypes.MsgCreateOracleScript:
		DecodeMsgCreateOracleScript(msg, detail)
	case *oracletypes.MsgEditDataSource:
		DecodeMsgEditDataSource(msg, detail)
	case *oracletypes.MsgEditOracleScript:
		DecodeMsgEditOracleScript(msg, detail)
	case *oracletypes.MsgActivate:
		DecodeMsgActivate(msg, detail)
	case *clienttypes.MsgCreateClient:
		DecodeMsgCreateClient(msg, detail)
	case *clienttypes.MsgUpdateClient:
		DecodeMsgUpdateClient(msg, detail)
	case *clienttypes.MsgUpgradeClient:
		DecodeMsgUpgradeClient(msg, detail)
	case *connectiontypes.MsgConnectionOpenInit:
		DecodeMsgConnectionOpenInit(msg, detail)
	case *connectiontypes.MsgConnectionOpenTry:
		DecodeMsgConnectionOpenTry(msg, detail)
	case *connectiontypes.MsgConnectionOpenAck:
		DecodeMsgConnectionOpenAck(msg, detail)
	case *connectiontypes.MsgConnectionOpenConfirm:
		DecodeMsgConnectionOpenConfirm(msg, detail)
	case *channeltypes.MsgChannelOpenInit:
		DecodeMsgChannelOpenInit(msg, detail)
	case *channeltypes.MsgChannelOpenTry:
		DecodeMsgChannelOpenTry(msg, detail)
	case *channeltypes.MsgChannelOpenAck:
		DecodeMsgChannelOpenAck(msg, detail)
	case *channeltypes.MsgChannelOpenConfirm:
		DecodeMsgChannelOpenConfirm(msg, detail)
	case *channeltypes.MsgChannelCloseInit:
		DecodeMsgChannelCloseInit(msg, detail)
	case *channeltypes.MsgChannelCloseConfirm:
		DecodeMsgChannelCloseConfirm(msg, detail)
	case *channeltypes.MsgRecvPacket:
		DecodeMsgRecvPacket(msg, detail)
	case *channeltypes.MsgAcknowledgement:
		DecodeMsgAcknowledgement(msg, detail)
	case *channeltypes.MsgTimeout:
		DecodeMsgTimeout(msg, detail)
	case *channeltypes.MsgTimeoutOnClose:
		DecodeMsgTimeoutOnClose(msg, detail)
	case *banktypes.MsgSend:
		DecodeMsgSend(msg, detail)
	case *banktypes.MsgMultiSend:
		DecodeMsgMultiSend(msg, detail)
	case *distrtypes.MsgSetWithdrawAddress:
		DecodeMsgSetWithdrawAddress(msg, detail)
	case *distrtypes.MsgWithdrawDelegatorReward:
		DecodeMsgWithdrawDelegatorReward(msg, detail)
	case *distrtypes.MsgWithdrawValidatorCommission:
		DecodeMsgWithdrawValidatorCommission(msg, detail)
	case *slashingtypes.MsgUnjail:
		DecodeMsgUnjail(msg, detail)
	case *transfertypes.MsgTransfer:
		DecodeMsgTransfer(msg, detail)
	case *govv1beta1.MsgSubmitProposal:
		DecodeV1beta1MsgSubmitProposal(msg, detail)
	case *govv1.MsgSubmitProposal:
		DecodeMsgSubmitProposal(msg, detail)
	case *govv1beta1.MsgDeposit:
		DecodeV1beta1MsgDeposit(msg, detail)
	case *govv1.MsgDeposit:
		DecodeMsgDeposit(msg, detail)
	case *govv1beta1.MsgVote:
		DecodeV1beta1MsgVote(msg, detail)
	case *govv1.MsgVote:
		DecodeMsgVote(msg, detail)
	case *govv1beta1.MsgVoteWeighted:
		DecodeV1beta1MsgVoteWeighted(msg, detail)
	case *govv1.MsgVoteWeighted:
		DecodeMsgVoteWeighted(msg, detail)
	case *stakingtypes.MsgCreateValidator:
		DecodeMsgCreateValidator(msg, detail)
	case *stakingtypes.MsgEditValidator:
		DecodeMsgEditValidator(msg, detail)
	case *stakingtypes.MsgDelegate:
		DecodeMsgDelegate(msg, detail)
	case *stakingtypes.MsgUndelegate:
		DecodeMsgUndelegate(msg, detail)
	case *stakingtypes.MsgBeginRedelegate:
		DecodeMsgBeginRedelegate(msg, detail)
	case *authz.MsgGrant:
		DecodeMsgGrant(msg, detail)
	case *authz.MsgRevoke:
		DecodeMsgRevoke(msg, detail)
	case *authz.MsgExec:
		DecodeMsgExec(msg, detail)
	case *feegranttypes.MsgGrantAllowance:
		DecodeMsgGrantAllowance(msg, detail)
	case *feegranttypes.MsgRevokeAllowance:
		DecodeMsgRevokeAllowance(msg, detail)
	case *feedstypes.MsgSubmitSignalPrices:
		DecodeFeedsMsgSubmitSignalPrices(msg, detail)
	case *feedstypes.MsgVote:
		DecodeFeedsMsgVote(msg, detail)
	case *feedstypes.MsgUpdateReferenceSourceConfig:
		DecodeFeedsMsgUpdateReferenceSourceConfig(msg, detail)
	case *feedstypes.MsgUpdateParams:
		DecodeFeedsMsgUpdateParams(msg, detail)
	case *bandtsstypes.MsgTransitionGroup:
		DecodeBandtssMsgTransitionGroup(msg, detail)
	case *bandtsstypes.MsgForceTransitionGroup:
		DecodeBandtssMsgForceTransitionGroup(msg, detail)
	case *bandtsstypes.MsgRequestSignature:
		DecodeBandtssMsgRequestSignature(msg, detail)
	case *bandtsstypes.MsgActivate:
		DecodeBandtssMsgActivate(msg, detail)
	case *bandtsstypes.MsgUpdateParams:
		DecodeBandtssMsgUpdateParams(msg, detail)
	case *tsstypes.MsgSubmitDKGRound1:
		DecodeTSSMsgSubmitDKGRound1(msg, detail)
	case *tsstypes.MsgSubmitDKGRound2:
		DecodeTSSMsgSubmitDKGRound2(msg, detail)
	case *tsstypes.MsgComplain:
		DecodeTSSMsgComplain(msg, detail)
	case *tsstypes.MsgConfirm:
		DecodeTSSMsgConfirm(msg, detail)
	case *tsstypes.MsgSubmitDEs:
		DecodeTSSMsgSubmitDEs(msg, detail)
	case *tsstypes.MsgSubmitSignature:
		DecodeTSSMsgSubmitSignature(msg, detail)
	case *tsstypes.MsgUpdateParams:
		DecodeTSSMsgUpdateParams(msg, detail)
	case *restaketypes.MsgStake:
		DecodeRestakeMsgStake(msg, detail)
	case *restaketypes.MsgUnstake:
		DecodeRestakeMsgUnstake(msg, detail)
	case *restaketypes.MsgUpdateParams:
		DecodeRestakeMsgUpdateParams(msg, detail)
	case *tunneltypes.MsgCreateTunnel:
		DecodeTunnelMsgCreateTunnel(msg, detail)
	case *tunneltypes.MsgUpdateSignalsAndInterval:
		DecodeTunnelMsgUpdateSignalsAndInterval(msg, detail)
	case *tunneltypes.MsgActivateTunnel:
		DecodeTunnelMsgActivateTunnel(msg, detail)
	case *tunneltypes.MsgDeactivateTunnel:
		DecodeTunnelMsgDeactivateTunnel(msg, detail)
	case *tunneltypes.MsgTriggerTunnel:
		DecodeTunnelMsgTriggerTunnel(msg, detail)
	case *tunneltypes.MsgDepositToTunnel:
		DecodeTunnelMsgDepositToTunnel(msg, detail)
	case *tunneltypes.MsgWithdrawFromTunnel:
		DecodeTunnelMsgWithdrawFromTunnel(msg, detail)
	case *tunneltypes.MsgUpdateParams:
		DecodeTunnelMsgUpdateParams(msg, detail)
	default:
		break
	}
}

func DecodeGrant(g authz.Grant) common.JsDict {
	authorization, _ := g.GetAuthorization()
	return common.JsDict{
		"authorization": authorization,
		"expiration":    g.Expiration,
	}
}

func DecodeMsgGrant(msg *authz.MsgGrant, detail common.JsDict) {
	detail["granter"] = msg.Granter
	detail["grantee"] = msg.Grantee
	detail["grant"] = DecodeGrant(msg.Grant)
}

func DecodeMsgRevoke(msg *authz.MsgRevoke, detail common.JsDict) {
	detail["granter"] = msg.Granter
	detail["grantee"] = msg.Grantee
	detail["msg_type_url"] = msg.MsgTypeUrl
}

func DecodeMsgExec(msg *authz.MsgExec, detail common.JsDict) {
	detail["grantee"] = msg.Grantee
	msgs, _ := msg.GetMessages()
	execMsgs := make([]common.JsDict, len(msgs))
	for i, msg := range msgs {
		detail := make(common.JsDict)
		DecodeMsg(msg, detail)
		execMsgs[i] = common.JsDict{
			"msg":  detail,
			"type": sdk.MsgTypeURL(msg),
		}
	}
	detail["msgs"] = execMsgs
}

func DecodeAllowance(allowance feegranttypes.FeeAllowanceI, detail common.JsDict) {
	switch allowance := allowance.(type) {
	case *feegranttypes.BasicAllowance:
		DecodeBasicAllowance(allowance, detail)
	case *feegranttypes.PeriodicAllowance:
		DecodePeriodicAllowance(allowance, detail)
	case *feegranttypes.AllowedMsgAllowance:
		DecodeAllowedMsgAllowance(allowance, detail)
	}
}

func DecodeBasicAllowance(allowance *feegranttypes.BasicAllowance, detail common.JsDict) {
	detail["spend_limit"] = allowance.GetSpendLimit()
	detail["expiration"] = allowance.GetExpiration()
	detail["type"] = "/cosmos.feegrant.v1beta1.BasicAllowance"
}

func DecodePeriodicAllowance(allowance *feegranttypes.PeriodicAllowance, detail common.JsDict) {
	detail["basic"] = allowance.GetBasic()
	detail["period"] = allowance.GetPeriod()
	detail["period_spend_limit"] = allowance.GetPeriodSpendLimit()
	detail["period_can_spend"] = allowance.GetPeriodCanSpend()
	detail["period_reset"] = allowance.GetPeriodReset()
	detail["type"] = "/cosmos.feegrant.v1beta1.PeriodicAllowance"
}

func DecodeAllowedMsgAllowance(allowance *feegranttypes.AllowedMsgAllowance, detail common.JsDict) {
	detail["allowed_messages"] = allowance.AllowedMessages
	detail["allowance"] = nil
	detail["type"] = "/cosmos.feegrant.v1beta1.AllowedMsgAllowance"
	sub_allowance, err := allowance.GetAllowance()
	if err == nil {
		allowance_detail := make(common.JsDict)
		DecodeAllowance(sub_allowance, allowance_detail)
		detail["allowance"] = allowance_detail
	}
}

func DecodeMsgGrantAllowance(msg *feegranttypes.MsgGrantAllowance, detail common.JsDict) {
	detail["granter"] = msg.GetGranter()
	detail["grantee"] = msg.GetGrantee()
	detail["allowance"] = nil
	allowance, err := msg.GetFeeAllowanceI()
	if err == nil {
		allowance_detail := make(common.JsDict)
		DecodeAllowance(allowance, allowance_detail)
		detail["allowance"] = allowance_detail
	}
}

func DecodeMsgRevokeAllowance(msg *feegranttypes.MsgRevokeAllowance, detail common.JsDict) {
	detail["granter"] = msg.GetGranter()
	detail["grantee"] = msg.GetGrantee()
}

func DecodeHeight(h clienttypes.Height) common.JsDict {
	return common.JsDict{
		"revision_number": h.GetRevisionNumber(),
		"revision_height": h.GetRevisionHeight(),
	}
}

func DecodePacket(packet channeltypes.Packet) common.JsDict {
	return common.JsDict{
		"sequence":            packet.GetSequence(),
		"source_port":         packet.GetSourcePort(),
		"source_channel":      packet.GetSourceChannel(),
		"destination_port":    packet.GetDestPort(),
		"destination_channel": packet.GetDestChannel(),
		"data":                packet.GetData(),
		"timeout_height": DecodeHeight(
			clienttypes.NewHeight(
				packet.GetTimeoutHeight().GetRevisionNumber(),
				packet.GetTimeoutHeight().GetRevisionHeight(),
			),
		),
		"timeout_timestamp": packet.GetTimeoutTimestamp(),
	}
}

func DecodeMsgRequestData(msg *oracletypes.MsgRequestData, detail common.JsDict) {
	detail["oracle_script_id"] = msg.GetOracleScriptID()
	detail["calldata"] = msg.GetCalldata()
	detail["ask_count"] = msg.GetAskCount()
	detail["min_count"] = msg.GetMinCount()
	detail["client_id"] = msg.GetClientID()
	detail["fee_limit"] = msg.GetFeeLimit()
	detail["prepare_gas"] = msg.GetPrepareGas()
	detail["execute_gas"] = msg.GetExecuteGas()
	detail["sender"] = msg.GetSender()
	detail["tss_encoder"] = msg.GetTSSEncoder()
}

func DecodeMsgReportData(msg *oracletypes.MsgReportData, detail common.JsDict) {
	detail["request_id"] = msg.GetRequestID()
	detail["raw_reports"] = msg.GetRawReports()
	detail["validator"] = msg.GetValidator()
}

func DecodeMsgCreateDataSource(msg *oracletypes.MsgCreateDataSource, detail common.JsDict) {
	detail["name"] = msg.GetName()
	detail["description"] = msg.GetDescription()
	detail["executable"] = msg.GetExecutable()
	detail["fee"] = msg.GetFee()
	detail["treasury"] = msg.GetTreasury()
	detail["owner"] = msg.GetOwner()
	detail["sender"] = msg.GetSender()
}

func DecodeMsgCreateOracleScript(msg *oracletypes.MsgCreateOracleScript, detail common.JsDict) {
	detail["name"] = msg.GetName()
	detail["description"] = msg.GetDescription()
	detail["schema"] = msg.GetSchema()
	detail["source_code_url"] = msg.GetSourceCodeURL()
	detail["code"] = msg.GetCode()
	detail["owner"] = msg.GetOwner()
	detail["sender"] = msg.GetSender()
}

func DecodeMsgEditDataSource(msg *oracletypes.MsgEditDataSource, detail common.JsDict) {
	detail["data_source_id"] = msg.GetDataSourceID()
	detail["name"] = msg.GetName()
	detail["description"] = msg.GetDescription()
	detail["executable"] = msg.GetExecutable()
	detail["fee"] = msg.GetFee()
	detail["treasury"] = msg.GetTreasury()
	detail["owner"] = msg.GetOwner()
	detail["sender"] = msg.GetSender()
}

func DecodeMsgEditOracleScript(msg *oracletypes.MsgEditOracleScript, detail common.JsDict) {
	detail["oracle_script_id"] = msg.GetOracleScriptID()
	detail["name"] = msg.GetName()
	detail["description"] = msg.GetDescription()
	detail["schema"] = msg.GetSchema()
	detail["source_code_url"] = msg.GetSourceCodeURL()
	detail["code"] = msg.GetCode()
	detail["owner"] = msg.GetOwner()
	detail["sender"] = msg.GetSender()
}

func DecodeMsgActivate(msg *oracletypes.MsgActivate, detail common.JsDict) {
	detail["validator"] = msg.GetValidator()
}

func DecodeMsgCreateClient(msg *clienttypes.MsgCreateClient, detail common.JsDict) {
	clientState, _ := clienttypes.UnpackClientState(msg.ClientState)
	consensusState, _ := clienttypes.UnpackConsensusState(msg.ConsensusState)

	detail["client_state"] = clientState
	detail["consensus_state"] = consensusState
	detail["signer"] = msg.Signer
}

func DecodeMsgSubmitProposal(msg *govv1.MsgSubmitProposal, detail common.JsDict) {
	detail["initial_deposit"] = msg.GetInitialDeposit()
	detail["proposer"] = msg.GetProposer()
	detail["metadata"] = msg.Metadata
	detail["title"] = msg.Title
	detail["summary"] = msg.Summary

	msgs, _ := msg.GetMsgs()
	messages := make([]common.JsDict, len(msgs))
	for i, m := range msgs {
		detail := make(common.JsDict)
		DecodeMsg(m, detail)
		messages[i] = common.JsDict{
			"msg":  detail,
			"type": sdk.MsgTypeURL(m),
		}
	}
	detail["messages"] = messages
}

func DecodeV1beta1MsgSubmitProposal(msg *govv1beta1.MsgSubmitProposal, detail common.JsDict) {
	detail["content"] = msg.GetContent()
	detail["initial_deposit"] = msg.GetInitialDeposit()
	detail["proposer"] = msg.Proposer
}

func DecodeMsgDeposit(msg *govv1.MsgDeposit, detail common.JsDict) {
	detail["proposal_id"] = msg.ProposalId
	detail["depositor"] = msg.Depositor
	detail["amount"] = msg.Amount
}

func DecodeV1beta1MsgDeposit(msg *govv1beta1.MsgDeposit, detail common.JsDict) {
	detail["proposal_id"] = msg.ProposalId
	detail["depositor"] = msg.Depositor
	detail["amount"] = msg.Amount
}

func DecodeMsgVote(msg *govv1.MsgVote, detail common.JsDict) {
	detail["proposal_id"] = msg.ProposalId
	detail["voter"] = msg.Voter
	detail["option"] = msg.Option
	detail["metadata"] = msg.Metadata
}

func DecodeV1beta1MsgVote(msg *govv1beta1.MsgVote, detail common.JsDict) {
	detail["proposal_id"] = msg.ProposalId
	detail["voter"] = msg.Voter
	detail["option"] = msg.Option
}

func DecodeMsgVoteWeighted(msg *govv1.MsgVoteWeighted, detail common.JsDict) {
	detail["proposal_id"] = msg.ProposalId
	detail["voter"] = msg.Voter
	detail["options"] = msg.Options
	detail["metadata"] = msg.Metadata
}

func DecodeV1beta1MsgVoteWeighted(msg *govv1beta1.MsgVoteWeighted, detail common.JsDict) {
	detail["proposal_id"] = msg.ProposalId
	detail["voter"] = msg.Voter
	detail["options"] = msg.Options
}

func DecodeMsgCreateValidator(msg *stakingtypes.MsgCreateValidator, detail common.JsDict) {
	pk, _ := msg.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	hexConsPubKey := hex.EncodeToString(pk.Bytes())

	detail["description"] = DecodeDescription(msg.Description)
	detail["commission"] = msg.Commission
	detail["min_self_delegation"] = msg.MinSelfDelegation
	detail["validator_address"] = msg.ValidatorAddress
	detail["pubkey"] = hexConsPubKey
	detail["value"] = msg.Value

	// delegatorAddress is deprecated. need to convert from validatorAddress
	addr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	detail["delegator_address"] = sdk.AccAddress(addr).String()
}

func DecodeMsgEditValidator(msg *stakingtypes.MsgEditValidator, detail common.JsDict) {
	detail["description"] = DecodeDescription(msg.Description)
	detail["validator_address"] = msg.ValidatorAddress
	detail["commission_rate"] = msg.CommissionRate
	detail["min_self_delegation"] = msg.MinSelfDelegation
}

func DecodeMsgDelegate(msg *stakingtypes.MsgDelegate, detail common.JsDict) {
	detail["delegator_address"] = msg.DelegatorAddress
	detail["validator_address"] = msg.ValidatorAddress
	detail["amount"] = msg.Amount
}

func DecodeMsgUndelegate(msg *stakingtypes.MsgUndelegate, detail common.JsDict) {
	detail["delegator_address"] = msg.DelegatorAddress
	detail["validator_address"] = msg.ValidatorAddress
	detail["amount"] = msg.Amount
}

func DecodeMsgBeginRedelegate(msg *stakingtypes.MsgBeginRedelegate, detail common.JsDict) {
	detail["delegator_address"] = msg.DelegatorAddress
	detail["validator_src_address"] = msg.ValidatorSrcAddress
	detail["validator_dst_address"] = msg.ValidatorDstAddress
	detail["amount"] = msg.Amount
}

func DecodeMsgUpdateClient(msg *clienttypes.MsgUpdateClient, detail common.JsDict) {
	cm, _ := clienttypes.UnpackClientMessage(msg.ClientMessage)
	detail["client_id"] = msg.ClientId
	detail["header"] = cm
	detail["signer"] = msg.Signer
}

func DecodeMsgUpgradeClient(msg *clienttypes.MsgUpgradeClient, detail common.JsDict) {
	clientState, _ := clienttypes.UnpackClientState(msg.ClientState)
	consensusState, _ := clienttypes.UnpackConsensusState(msg.ConsensusState)
	detail["client_id"] = msg.ClientId
	detail["client_state"] = clientState
	detail["consensus_state"] = consensusState
	detail["proof_upgrade_client"] = msg.ProofUpgradeClient
	detail["proof_upgrade_consensus_state"] = msg.ProofUpgradeConsensusState
	detail["signer"] = msg.Signer
}

// MsgSubmitMisbehaviour is deprecated but still use able.
//
//nolint:staticcheck
func DecodeMsgSubmitMisbehaviour(msg *clienttypes.MsgSubmitMisbehaviour, detail common.JsDict) {
	misbehaviour, _ := clienttypes.UnpackClientMessage(msg.Misbehaviour)
	detail["client_id"] = msg.ClientId
	detail["misbehaviour"] = misbehaviour
	detail["signer"] = msg.Signer
}

func DecodeMsgConnectionOpenInit(msg *connectiontypes.MsgConnectionOpenInit, detail common.JsDict) {
	detail["client_id"] = msg.ClientId
	detail["counterparty"] = msg.Counterparty
	detail["version"] = msg.Version
	detail["delay_period"] = msg.DelayPeriod
	detail["signer"] = msg.Signer
}

func DecodeMsgConnectionOpenTry(msg *connectiontypes.MsgConnectionOpenTry, detail common.JsDict) {
	clientState, _ := clienttypes.UnpackClientState(msg.ClientState)
	detail["client_id"] = msg.ClientId
	detail["previous_connection_id"] = ""
	detail["client_state"] = clientState
	detail["counterparty"] = msg.Counterparty
	detail["delay_period"] = msg.DelayPeriod
	detail["counterparty_versions"] = msg.CounterpartyVersions
	detail["proof_height"] = DecodeHeight(msg.ProofHeight)
	detail["proof_init"] = msg.ProofInit
	detail["proof_client"] = msg.ProofClient
	detail["proof_consensus"] = msg.ProofConsensus
	detail["consensus_height"] = DecodeHeight(msg.ConsensusHeight)
	detail["signer"] = msg.Signer
}

func DecodeMsgConnectionOpenAck(msg *connectiontypes.MsgConnectionOpenAck, detail common.JsDict) {
	clientState, _ := clienttypes.UnpackClientState(msg.ClientState)
	detail["connection_id"] = msg.ConnectionId
	detail["counterparty_connection_id"] = msg.CounterpartyConnectionId
	detail["version"] = msg.Version
	detail["client_state"] = clientState
	detail["proof_height"] = DecodeHeight(msg.ProofHeight)
	detail["proof_try"] = msg.ProofTry
	detail["proof_client"] = msg.ProofClient
	detail["proof_consensus"] = msg.ProofConsensus
	detail["consensus_height"] = DecodeHeight(msg.ConsensusHeight)
	detail["signer"] = msg.Signer
}

func DecodeMsgConnectionOpenConfirm(msg *connectiontypes.MsgConnectionOpenConfirm, detail common.JsDict) {
	detail["connection_id"] = msg.ConnectionId
	detail["proof_ack"] = msg.ProofAck
	detail["proof_height"] = DecodeHeight(msg.ProofHeight)
	detail["signer"] = msg.Signer
}

func DecodeMsgChannelOpenInit(msg *channeltypes.MsgChannelOpenInit, detail common.JsDict) {
	detail["port_id"] = msg.PortId
	detail["channel"] = msg.Channel
	detail["signer"] = msg.Signer
}

func DecodeMsgChannelOpenTry(msg *channeltypes.MsgChannelOpenTry, detail common.JsDict) {
	detail["port_id"] = msg.PortId
	detail["previous_channel_id"] = ""
	detail["channel"] = msg.Channel
	detail["counterparty_version"] = msg.CounterpartyVersion
	detail["proof_init"] = msg.ProofInit
	detail["proof_height"] = DecodeHeight(msg.ProofHeight)
	detail["signer"] = msg.Signer
}

func DecodeMsgChannelOpenAck(msg *channeltypes.MsgChannelOpenAck, detail common.JsDict) {
	detail["port_id"] = msg.PortId
	detail["channel_id"] = msg.ChannelId
	detail["counterparty_channel_id"] = msg.CounterpartyChannelId
	detail["counterparty_version"] = msg.CounterpartyVersion
	detail["proof_try"] = msg.ProofTry
	detail["proof_height"] = DecodeHeight(msg.ProofHeight)
	detail["signer"] = msg.Signer
}

func DecodeMsgChannelOpenConfirm(msg *channeltypes.MsgChannelOpenConfirm, detail common.JsDict) {
	detail["port_id"] = msg.PortId
	detail["channel_id"] = msg.ChannelId
	detail["proof_ack"] = msg.ProofAck
	detail["proof_height"] = DecodeHeight(msg.ProofHeight)
	detail["signer"] = msg.Signer
}

func DecodeMsgChannelCloseInit(msg *channeltypes.MsgChannelCloseInit, detail common.JsDict) {
	detail["port_id"] = msg.PortId
	detail["channel_id"] = msg.ChannelId
	detail["signer"] = msg.Signer
}

func DecodeMsgChannelCloseConfirm(msg *channeltypes.MsgChannelCloseConfirm, detail common.JsDict) {
	detail["port_id"] = msg.PortId
	detail["channel_id"] = msg.ChannelId
	detail["proof_init"] = msg.ProofInit
	detail["proof_height"] = DecodeHeight(msg.ProofHeight)
	detail["signer"] = msg.Signer
}

func DecodeMsgRecvPacket(msg *channeltypes.MsgRecvPacket, detail common.JsDict) {
	detail["packet"] = DecodePacket(msg.Packet)
	detail["proof_commitment"] = msg.ProofCommitment
	detail["proof_height"] = DecodeHeight(msg.ProofHeight)
	detail["signer"] = msg.Signer
}

func DecodeMsgAcknowledgement(msg *channeltypes.MsgAcknowledgement, detail common.JsDict) {
	detail["packet"] = DecodePacket(msg.Packet)
	detail["acknowledgement"] = msg.Acknowledgement
	detail["proof_acked"] = msg.ProofAcked
	detail["proof_height"] = DecodeHeight(msg.ProofHeight)
	detail["signer"] = msg.Signer
}

func DecodeMsgTimeout(msg *channeltypes.MsgTimeout, detail common.JsDict) {
	detail["packet"] = DecodePacket(msg.Packet)
	detail["proof_unreceived"] = msg.ProofUnreceived
	detail["proof_height"] = DecodeHeight(msg.ProofHeight)
	detail["next_sequence_recv"] = msg.NextSequenceRecv
	detail["signer"] = msg.Signer
}

func DecodeMsgTimeoutOnClose(msg *channeltypes.MsgTimeoutOnClose, detail common.JsDict) {
	detail["packet"] = DecodePacket(msg.Packet)
	detail["proof_unreceived"] = msg.ProofUnreceived
	detail["proof_close"] = msg.ProofClose
	detail["proof_height"] = DecodeHeight(msg.ProofHeight)
	detail["next_sequence_recv"] = msg.NextSequenceRecv
	detail["signer"] = msg.Signer
}

func DecodeMsgSend(msg *banktypes.MsgSend, detail common.JsDict) {
	detail["from_address"] = msg.FromAddress
	detail["to_address"] = msg.ToAddress
	detail["amount"] = msg.Amount
}

func DecodeMsgMultiSend(msg *banktypes.MsgMultiSend, detail common.JsDict) {
	detail["inputs"] = msg.Inputs
	detail["outputs"] = msg.Outputs
}

func DecodeMsgSetWithdrawAddress(msg *distrtypes.MsgSetWithdrawAddress, detail common.JsDict) {
	detail["delegator_address"] = msg.DelegatorAddress
	detail["withdraw_address"] = msg.WithdrawAddress
}

func DecodeMsgWithdrawDelegatorReward(msg *distrtypes.MsgWithdrawDelegatorReward, detail common.JsDict) {
	detail["delegator_address"] = msg.DelegatorAddress
	detail["validator_address"] = msg.ValidatorAddress
}

func DecodeMsgWithdrawValidatorCommission(msg *distrtypes.MsgWithdrawValidatorCommission, detail common.JsDict) {
	detail["validator_address"] = msg.ValidatorAddress
}

func DecodeMsgUnjail(msg *slashingtypes.MsgUnjail, detail common.JsDict) {
	detail["validator_address"] = msg.ValidatorAddr
}

func DecodeMsgTransfer(msg *transfertypes.MsgTransfer, detail common.JsDict) {
	detail["source_port"] = msg.SourcePort
	detail["source_channel"] = msg.SourceChannel
	detail["token"] = msg.Token
	detail["sender"] = msg.Sender
	detail["receiver"] = msg.Receiver
	detail["timeout_height"] = DecodeHeight(msg.TimeoutHeight)
	detail["timeout_timestamp"] = msg.TimeoutTimestamp
}

func DecodeDescription(des stakingtypes.Description) common.JsDict {
	return common.JsDict{
		"details":          des.GetDetails(),
		"identity":         des.GetIdentity(),
		"moniker":          des.GetMoniker(),
		"security_contact": des.GetSecurityContact(),
		"website":          des.GetWebsite(),
	}
}

func DecodeFeedsMsgSubmitSignalPrices(msg *feedstypes.MsgSubmitSignalPrices, detail common.JsDict) {
	detail["validator"] = msg.GetValidator()
	detail["timestamp"] = msg.GetTimestamp()
	detail["signal_prices"] = msg.GetSignalPrices()
}

func DecodeFeedsMsgVote(msg *feedstypes.MsgVote, detail common.JsDict) {
	detail["voter"] = msg.GetVoter()
	detail["signals"] = msg.GetSignals()
}

func DecodeFeedsMsgUpdateReferenceSourceConfig(msg *feedstypes.MsgUpdateReferenceSourceConfig, detail common.JsDict) {
	detail["admin"] = msg.GetAdmin()
	detail["reference_source_config"] = msg.GetReferenceSourceConfig()
}

func DecodeFeedsMsgUpdateParams(msg *feedstypes.MsgUpdateParams, detail common.JsDict) {
	detail["authority"] = msg.GetAuthority()
	detail["params"] = msg.GetParams()
}

func DecodeBandtssMsgTransitionGroup(msg *bandtsstypes.MsgTransitionGroup, detail common.JsDict) {
	detail["members"] = msg.Members
	detail["threshold"] = msg.Threshold
	detail["exec_time"] = msg.ExecTime.UnixNano()
	detail["authority"] = msg.Authority
}

func DecodeBandtssMsgForceTransitionGroup(msg *bandtsstypes.MsgForceTransitionGroup, detail common.JsDict) {
	detail["incoming_group_id"] = msg.IncomingGroupID
	detail["exec_time"] = msg.ExecTime.UnixNano()
	detail["authority"] = msg.Authority
}

func DecodeBandtssMsgRequestSignature(msg *bandtsstypes.MsgRequestSignature, detail common.JsDict) {
	detail["content_type"] = msg.Content.TypeUrl
	detail["content"] = msg.Content.GetCachedValue()
	detail["memo"] = msg.Memo
	detail["fee_limit"] = msg.FeeLimit
	detail["sender"] = msg.Sender
}

func DecodeBandtssMsgActivate(msg *bandtsstypes.MsgActivate, detail common.JsDict) {
	detail["sender"] = msg.Sender
	detail["group_id"] = msg.GroupID
}

func DecodeBandtssMsgUpdateParams(msg *bandtsstypes.MsgUpdateParams, detail common.JsDict) {
	detail["params"] = msg.GetParams()
	detail["authority"] = msg.GetAuthority()
}

func DecodeTSSMsgSubmitDKGRound1(msg *tsstypes.MsgSubmitDKGRound1, detail common.JsDict) {
	detail["group_id"] = msg.GroupID
	detail["round1_info"] = msg.Round1Info
	detail["sender"] = msg.Sender
}

func DecodeTSSMsgSubmitDKGRound2(msg *tsstypes.MsgSubmitDKGRound2, detail common.JsDict) {
	detail["group_id"] = msg.GroupID
	detail["round2_info"] = msg.Round2Info
	detail["sender"] = msg.Sender
}

func DecodeTSSMsgComplain(msg *tsstypes.MsgComplain, detail common.JsDict) {
	detail["group_id"] = msg.GroupID
	detail["complaints"] = msg.Complaints
	detail["sender"] = msg.Sender
}

func DecodeTSSMsgConfirm(msg *tsstypes.MsgConfirm, detail common.JsDict) {
	detail["group_id"] = msg.GroupID
	detail["member_id"] = msg.MemberID
	detail["own_pub_key_sig"] = msg.OwnPubKeySig
	detail["sender"] = msg.Sender
}

func DecodeTSSMsgSubmitDEs(msg *tsstypes.MsgSubmitDEs, detail common.JsDict) {
	detail["des"] = msg.DEs
	detail["sender"] = msg.Sender
}

func DecodeTSSMsgSubmitSignature(msg *tsstypes.MsgSubmitSignature, detail common.JsDict) {
	detail["signing_id"] = msg.SigningID
	detail["member_id"] = msg.MemberID
	detail["signature"] = msg.Signature
	detail["signer"] = msg.Signer
}

func DecodeTSSMsgUpdateParams(msg *tsstypes.MsgUpdateParams, detail common.JsDict) {
	detail["params"] = msg.GetParams()
	detail["authority"] = msg.GetAuthority()
}

func DecodeRestakeMsgStake(msg *restaketypes.MsgStake, detail common.JsDict) {
	detail["staker_address"] = msg.StakerAddress
	detail["coins"] = msg.GetCoins()
}

func DecodeRestakeMsgUnstake(msg *restaketypes.MsgUnstake, detail common.JsDict) {
	detail["staker_address"] = msg.StakerAddress
	detail["coins"] = msg.GetCoins()
}

func DecodeRestakeMsgUpdateParams(msg *restaketypes.MsgUpdateParams, detail common.JsDict) {
	detail["authority"] = msg.Authority
	detail["params"] = msg.GetParams()
}

func DecodeTunnelMsgCreateTunnel(msg *tunneltypes.MsgCreateTunnel, detail common.JsDict) {
	detail["signal_deviations"] = msg.GetSignalDeviations()
	detail["interval"] = msg.Interval
	detail["route_type"] = msg.GetRoute().TypeUrl
	detail["route"] = msg.GetRoute().GetCachedValue()
	detail["initial_deposit"] = msg.InitialDeposit
	detail["creator"] = msg.Creator
}

func DecodeTunnelMsgUpdateSignalsAndInterval(msg *tunneltypes.MsgUpdateSignalsAndInterval, detail common.JsDict) {
	detail["tunnel_id"] = msg.TunnelID
	detail["signal_deviations"] = msg.GetSignalDeviations()
	detail["interval"] = msg.Interval
	detail["creator"] = msg.Creator
}

func DecodeTunnelMsgActivateTunnel(msg *tunneltypes.MsgActivateTunnel, detail common.JsDict) {
	detail["tunnel_id"] = msg.TunnelID
	detail["creator"] = msg.Creator
}

func DecodeTunnelMsgDeactivateTunnel(msg *tunneltypes.MsgDeactivateTunnel, detail common.JsDict) {
	detail["tunnel_id"] = msg.TunnelID
	detail["creator"] = msg.Creator
}

func DecodeTunnelMsgTriggerTunnel(msg *tunneltypes.MsgTriggerTunnel, detail common.JsDict) {
	detail["tunnel_id"] = msg.TunnelID
	detail["creator"] = msg.Creator
}

func DecodeTunnelMsgDepositToTunnel(msg *tunneltypes.MsgDepositToTunnel, detail common.JsDict) {
	detail["tunnel_id"] = msg.TunnelID
	detail["amount"] = msg.GetAmount()
	detail["depositor"] = msg.Depositor
}

func DecodeTunnelMsgWithdrawFromTunnel(msg *tunneltypes.MsgWithdrawFromTunnel, detail common.JsDict) {
	detail["tunnel_id"] = msg.TunnelID
	detail["amount"] = msg.GetAmount()
	detail["withdrawer"] = msg.Withdrawer
}

func DecodeTunnelMsgUpdateParams(msg *tunneltypes.MsgUpdateParams, detail common.JsDict) {
	detail["authority"] = msg.Authority
	detail["params"] = msg.GetParams()
}
