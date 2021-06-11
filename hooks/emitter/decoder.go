package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/hooks/common"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"

	oracletypes "github.com/bandprotocol/chain/x/oracle/types"
)

func (h *Hook) decodeMsg(ctx sdk.Context, msg sdk.Msg, detail common.JsDict) {
	switch msg := msg.(type) {
	case *oracletypes.MsgRequestData:
		decodeMsgRequestData(msg, detail)
	case *oracletypes.MsgReportData:
		decodeMsgReportData(msg, detail)
	case *oracletypes.MsgCreateDataSource:
		decodeMsgCreateDataSource(msg, detail)
	case *oracletypes.MsgCreateOracleScript:
		decodeMsgCreateOracleScript(msg, detail)
	case *oracletypes.MsgEditDataSource:
		decodeMsgEditDataSource(msg, detail)
	case *oracletypes.MsgEditOracleScript:
		decodeMsgEditOracleScript(msg, detail)
	case *oracletypes.MsgAddReporter:
		decodeMsgAddReporter(msg, detail)
	case *oracletypes.MsgRemoveReporter:
		decodeMsgRemoveReporter(msg, detail)
	case *oracletypes.MsgActivate:
		decodeMsgActivate(msg, detail)
	case *clienttypes.MsgCreateClient:
		decodeMsgCreateClient(msg, detail)
	default:
		break
	}
}

func decodeMsgRequestData(msg *oracletypes.MsgRequestData, detail common.JsDict) {
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

func decodeMsgReportData(msg *oracletypes.MsgReportData, detail common.JsDict) {
	detail["request_id"] = msg.GetRequestID()
	detail["raw_reports"] = msg.GetRawReports()
	detail["validator"] = msg.GetValidator()
	detail["reporter"] = msg.GetReporter()
}

func decodeMsgCreateDataSource(msg *oracletypes.MsgCreateDataSource, detail common.JsDict) {
	detail["name"] = msg.GetName()
	detail["description"] = msg.GetDescription()
	detail["executable"] = msg.GetExecutable()
	detail["fee"] = msg.GetFee()
	detail["treasury"] = msg.GetTreasury()
	detail["owner"] = msg.GetOwner()
	detail["sender"] = msg.GetSender()
}

func decodeMsgCreateOracleScript(msg *oracletypes.MsgCreateOracleScript, detail common.JsDict) {
	detail["name"] = msg.GetName()
	detail["description"] = msg.GetDescription()
	detail["schema"] = msg.GetSchema()
	detail["source_code_url"] = msg.GetSourceCodeURL()
	detail["code"] = msg.GetCode()
	detail["owner"] = msg.GetOwner()
	detail["sender"] = msg.GetSender()
}

func decodeMsgEditDataSource(msg *oracletypes.MsgEditDataSource, detail common.JsDict) {
	detail["data_source_id"] = msg.GetDataSourceID()
	detail["name"] = msg.GetName()
	detail["description"] = msg.GetDescription()
	detail["executable"] = msg.GetExecutable()
	detail["fee"] = msg.GetFee()
	detail["treasury"] = msg.GetTreasury()
	detail["owner"] = msg.GetOwner()
	detail["sender"] = msg.GetSender()
}

func decodeMsgEditOracleScript(msg *oracletypes.MsgEditOracleScript, detail common.JsDict) {
	detail["oracle_script_id"] = msg.GetOracleScriptID()
	detail["name"] = msg.GetName()
	detail["description"] = msg.GetDescription()
	detail["schema"] = msg.GetSchema()
	detail["source_code_url"] = msg.GetSourceCodeURL()
	detail["code"] = msg.GetCode()
	detail["owner"] = msg.GetOwner()
	detail["sender"] = msg.GetSender()
}

func decodeMsgAddReporter(msg *oracletypes.MsgAddReporter, detail common.JsDict) {
	detail["validator"] = msg.GetValidator()
	detail["reporter"] = msg.GetReporter()
}

func decodeMsgRemoveReporter(msg *oracletypes.MsgRemoveReporter, detail common.JsDict) {
	detail["validator"] = msg.GetValidator()
	detail["reporter"] = msg.GetReporter()
}

func decodeMsgActivate(msg *oracletypes.MsgActivate, detail common.JsDict) {
	detail["validator"] = msg.GetValidator()
}

func decodeMsgCreateClient(msg *clienttypes.MsgCreateClient, detail common.JsDict) {
	clientState, _ := clienttypes.UnpackClientState(msg.ClientState)
	consensusState, _ := clienttypes.UnpackConsensusState(msg.ConsensusState)

	detail["client_state"] = clientState
	detail["consensus_state"] = consensusState
	detail["signer"] = msg.Signer
}
