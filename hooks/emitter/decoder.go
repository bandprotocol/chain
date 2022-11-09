package emitter

import (
	"encoding/hex"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v3/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	"github.com/bandprotocol/chain/v2/hooks/common"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
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
	case *clienttypes.MsgSubmitMisbehaviour:
		DecodeMsgSubmitMisbehaviour(msg, detail)
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
	case *govtypes.MsgSubmitProposal:
		DecodeMsgSubmitProposal(msg, detail)
	case *govtypes.MsgDeposit:
		DecodeMsgDeposit(msg, detail)
	case *govtypes.MsgVote:
		DecodeMsgVote(msg, detail)
	case *govtypes.MsgVoteWeighted:
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
	default:
		break
	}
}

func DecodeGrant(g authz.Grant) common.JsDict {
	return common.JsDict{
		"authorization": g.GetAuthorization(),
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

func DecodeMsgSubmitProposal(msg *govtypes.MsgSubmitProposal, detail common.JsDict) {
	detail["content"] = msg.GetContent()
	detail["initial_deposit"] = msg.GetInitialDeposit()
	detail["proposer"] = msg.GetProposer()
}

func DecodeMsgDeposit(msg *govtypes.MsgDeposit, detail common.JsDict) {
	detail["proposal_id"] = msg.ProposalId
	detail["depositor"] = msg.Depositor
	detail["amount"] = msg.Amount
}

func DecodeMsgVote(msg *govtypes.MsgVote, detail common.JsDict) {
	detail["proposal_id"] = msg.ProposalId
	detail["voter"] = msg.Voter
	detail["option"] = msg.Option
}

func DecodeMsgVoteWeighted(msg *govtypes.MsgVoteWeighted, detail common.JsDict) {
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
	detail["delegator_address"] = msg.DelegatorAddress
	detail["validator_address"] = msg.ValidatorAddress
	detail["pubkey"] = hexConsPubKey
	detail["value"] = msg.Value
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
	header, _ := clienttypes.UnpackHeader(msg.Header)
	detail["client_id"] = msg.ClientId
	detail["header"] = header
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

func DecodeMsgSubmitMisbehaviour(msg *clienttypes.MsgSubmitMisbehaviour, detail common.JsDict) {
	misbehaviour, _ := clienttypes.UnpackMisbehaviour(msg.Misbehaviour)
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
	detail["previous_connection_id"] = msg.PreviousConnectionId
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
	detail["previous_channel_id"] = msg.PreviousChannelId
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
