package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/hooks/common"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"

	oracletypes "github.com/bandprotocol/chain/x/oracle/types"
)

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
	case *clienttypes.MsgCreateClient:
		decodeMsgCreateClient(msg, msgJson)
	default:
		break
	}
}

func decodeMsgRequestData(msg *oracletypes.MsgRequestData, msgJson common.JsDict) {
	msgJson["oracle_script_id"] = msg.GetOracleScriptID()
	msgJson["calldata"] = msg.GetCalldata()
	msgJson["ask_count"] = msg.GetAskCount()
	msgJson["min_count"] = msg.GetMinCount()
	msgJson["client_id"] = msg.GetClientID()
	msgJson["fee_limit"] = msg.GetFeeLimit()
	msgJson["prepare_gas"] = msg.GetPrepareGas()
	msgJson["execute_gas"] = msg.GetExecuteGas()
	msgJson["sender"] = msg.GetSender()
}

func decodeMsgReportData(msg *oracletypes.MsgReportData, msgJson common.JsDict) {
	msgJson["request_id"] = msg.GetRequestID()
	msgJson["raw_reports"] = msg.GetRawReports()
	msgJson["validator"] = msg.GetValidator()
	msgJson["reporter"] = msg.GetReporter()
}

func decodeMsgCreateDataSource(msg *oracletypes.MsgCreateDataSource, msgJson common.JsDict) {
	msgJson["name"] = msg.GetName()
	msgJson["description"] = msg.GetDescription()
	msgJson["executable"] = msg.GetExecutable()
	msgJson["fee"] = msg.GetFee()
	msgJson["treasury"] = msg.GetTreasury()
	msgJson["owner"] = msg.GetOwner()
	msgJson["sender"] = msg.GetSender()
}

func decodeMsgCreateOracleScript(msg *oracletypes.MsgCreateOracleScript, msgJson common.JsDict) {
	msgJson["name"] = msg.GetName()
	msgJson["description"] = msg.GetDescription()
	msgJson["schema"] = msg.GetSchema()
	msgJson["source_code_url"] = msg.GetSourceCodeURL()
	msgJson["code"] = msg.GetCode()
	msgJson["owner"] = msg.GetOwner()
	msgJson["sender"] = msg.GetSender()
}

func decodeMsgEditDataSource(msg *oracletypes.MsgEditDataSource, msgJson common.JsDict) {
	msgJson["data_source_id"] = msg.GetDataSourceID()
	msgJson["name"] = msg.GetName()
	msgJson["description"] = msg.GetDescription()
	msgJson["executable"] = msg.GetExecutable()
	msgJson["fee"] = msg.GetFee()
	msgJson["treasury"] = msg.GetTreasury()
	msgJson["owner"] = msg.GetOwner()
	msgJson["sender"] = msg.GetSender()
}

func decodeMsgEditOracleScript(msg *oracletypes.MsgEditOracleScript, msgJson common.JsDict) {
	msgJson["oracle_script_id"] = msg.GetOracleScriptID()
	msgJson["name"] = msg.GetName()
	msgJson["description"] = msg.GetDescription()
	msgJson["schema"] = msg.GetSchema()
	msgJson["source_code_url"] = msg.GetSourceCodeURL()
	msgJson["code"] = msg.GetCode()
	msgJson["owner"] = msg.GetOwner()
	msgJson["sender"] = msg.GetSender()
}

func decodeMsgAddReporter(msg *oracletypes.MsgAddReporter, msgJson common.JsDict) {
	msgJson["validator"] = msg.GetValidator()
	msgJson["reporter"] = msg.GetReporter()
}

func decodeMsgRemoveReporter(msg *oracletypes.MsgRemoveReporter, msgJson common.JsDict) {
	msgJson["validator"] = msg.GetValidator()
	msgJson["reporter"] = msg.GetReporter()
}

func decodeMsgActivate(msg *oracletypes.MsgActivate, msgJson common.JsDict) {
	msgJson["validator"] = msg.GetValidator()
}

func decodeMsgCreateClient(msg *clienttypes.MsgCreateClient, msgJson common.JsDict) {
	clientState, _ := clienttypes.UnpackClientState(msg.ClientState)
	consensusState, _ := clienttypes.UnpackConsensusState(msg.ConsensusState)

	msgJson["client_state"] = clientState
	msgJson["consensus_state"] = consensusState
	msgJson["signer"] = msg.Signer
}
