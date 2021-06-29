package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/hooks/common"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

func parseBytes(b []byte) []byte {
	if len(b) == 0 {
		return []byte{}
	}
	return b
}

func (h *Hook) emitOracleModule(ctx sdk.Context) {
	dataSources := h.oracleKeeper.GetAllDataSources(ctx)
	for idx, ds := range dataSources {
		id := types.DataSourceID(idx + 1)
		h.emitSetDataSource(id, ds, nil)
	}
	oracleScripts := h.oracleKeeper.GetAllOracleScripts(ctx)
	for idx, os := range oracleScripts {
		id := types.OracleScriptID(idx + 1)
		h.emitSetOracleScript(id, os, nil)
	}
	rqCount := h.oracleKeeper.GetRequestCount(ctx)
	for rid := types.RequestID(1); rid <= types.RequestID(rqCount); rid++ {
		// TODO: revisit must get request
		req := h.oracleKeeper.MustGetRequest(ctx, rid)
		h.Write("NEW_REQUEST", common.JsDict{
			"id":               rid,
			"oracle_script_id": req.OracleScriptID,
			"calldata":         parseBytes(req.Calldata),
			"ask_count":        len(req.RequestedValidators),
			"min_count":        req.MinCount,
			"tx_hash":          nil,
			"client_id":        req.ClientID,
			"resolve_status":   types.RESOLVE_STATUS_OPEN,
			"execute_gas":      req.ExecuteGas,
		})
		if h.oracleKeeper.HasResult(ctx, rid) {
			h.emitUpdateResult(ctx, rid, "")
		}
		h.emitRawRequestAndValRequest(rid, req)
		reps := h.oracleKeeper.GetReports(ctx, rid)
		for _, rep := range reps {
			h.emitReportAndRawReport(nil, rid, sdk.ValAddress(rep.Validator), nil, rep.RawReports)
		}
	}
}

func (h *Hook) emitSetDataSource(id types.DataSourceID, ds types.DataSource, txHash []byte) {
	h.Write("SET_DATA_SOURCE", common.JsDict{
		"id":          id,
		"name":        ds.Name,
		"description": ds.Description,
		"owner":       ds.Owner,
		"executable":  h.oracleKeeper.GetFile(ds.Filename),
		"fee":         ds.Fee.String(),
		"treasury":    ds.Treasury,
		"tx_hash":     txHash,
	})
}

func (h *Hook) emitSetOracleScript(id types.OracleScriptID, os types.OracleScript, txHash []byte) {
	h.Write("SET_ORACLE_SCRIPT", common.JsDict{
		"id":              id,
		"name":            os.Name,
		"description":     os.Description,
		"owner":           os.Owner,
		"schema":          os.Schema,
		"codehash":        os.Filename,
		"source_code_url": os.SourceCodeURL,
		"tx_hash":         txHash,
	})
}

func (h *Hook) emitHistoricalValidatorStatus(ctx sdk.Context, operatorAddress sdk.ValAddress) {
	status := h.oracleKeeper.GetValidatorStatus(ctx, operatorAddress).IsActive
	h.Write("SET_HISTORICAL_VALIDATOR_STATUS", common.JsDict{
		"operator_address": operatorAddress,
		"status":           status,
		"timestamp":        ctx.BlockTime().UnixNano(),
	})
}

func (h *Hook) emitRawRequestAndValRequest(requestID types.RequestID, req types.Request) {
	for _, raw := range req.RawRequests {
		h.Write("NEW_RAW_REQUEST", common.JsDict{
			"request_id":     requestID,
			"external_id":    raw.ExternalID,
			"data_source_id": raw.DataSourceID,
			"calldata":       parseBytes(raw.Calldata),
		})
	}
	for _, val := range req.RequestedValidators {
		h.Write("NEW_VAL_REQUEST", common.JsDict{
			"request_id": requestID,
			"validator":  val,
		})
	}
}

func (app *Hook) emitReportAndRawReport(
	txHash []byte, rid types.RequestID, validator sdk.ValAddress, reporter sdk.AccAddress, rawReports []types.RawReport,
) {
	app.Write("NEW_REPORT", common.JsDict{
		"tx_hash":    txHash,
		"request_id": rid,
		"validator":  validator.String(),
		"reporter":   reporter.String(),
	})
	for _, data := range rawReports {
		app.Write("NEW_RAW_REPORT", common.JsDict{
			"request_id":  rid,
			"validator":   validator.String(),
			"external_id": data.ExternalID,
			"data":        parseBytes(data.Data),
			"exit_code":   data.ExitCode,
		})
	}
}

func (h *Hook) emitUpdateResult(ctx sdk.Context, id types.RequestID, reason string) {
	result := h.oracleKeeper.MustGetResult(ctx, id)
	h.Write("UPDATE_REQUEST", common.JsDict{
		"id":             id,
		"request_time":   result.RequestTime,
		"resolve_time":   result.ResolveTime,
		"resolve_status": result.ResolveStatus,
		"resolve_height": ctx.BlockHeight(),
		"reason":         reason,
		"result":         parseBytes(result.Result),
	})
}

// handleMsgRequestData implements emitter handler for MsgRequestData.
func (h *Hook) handleMsgRequestData(
	ctx sdk.Context, txHash []byte, msg *types.MsgRequestData, evMap common.EvMap, detail common.JsDict,
) {
	id := types.RequestID(common.Atoi(evMap[types.EventTypeRequest+"."+types.AttributeKeyID][0]))
	req := h.oracleKeeper.MustGetRequest(ctx, id)
	h.Write("NEW_REQUEST", common.JsDict{
		"id":               id,
		"tx_hash":          txHash,
		"oracle_script_id": msg.OracleScriptID,
		"calldata":         parseBytes(msg.Calldata),
		"ask_count":        msg.AskCount,
		"min_count":        msg.MinCount,
		"sender":           msg.Sender,
		"client_id":        msg.ClientID,
		"resolve_status":   types.RESOLVE_STATUS_OPEN,
		"timestamp":        ctx.BlockTime().UnixNano(),
		"prepare_gas":      msg.PrepareGas,
		"execute_gas":      msg.ExecuteGas,
		"fee_limit":        msg.FeeLimit.String(),
		"total_fees":       evMap[types.EventTypeRequest+"."+types.AttributeKeyTotalFees][0],
		"is_ibc":           req.IBCChannel != nil,
	})
	h.emitRawRequestAndValRequest(id, req)
	os := h.oracleKeeper.MustGetOracleScript(ctx, msg.OracleScriptID)
	detail["id"] = id
	detail["name"] = os.Name
	detail["schema"] = os.Schema
}

// handleMsgReportData implements emitter handler for MsgReportData.
func (h *Hook) handleMsgReportData(
	ctx sdk.Context, txHash []byte, msg *types.MsgReportData, evMap common.EvMap,
) {
	val, _ := sdk.ValAddressFromBech32(msg.Validator)
	rep, _ := sdk.AccAddressFromBech32(msg.Reporter)
	h.emitReportAndRawReport(txHash, msg.RequestID, val, rep, msg.RawReports)
}

// handleMsgCreateDataSource implements emitter handler for MsgCreateDataSource.
func (h *Hook) handleMsgCreateDataSource(
	ctx sdk.Context, txHash []byte, evMap common.EvMap, detail common.JsDict,
) {
	id := types.DataSourceID(common.Atoi(evMap[types.EventTypeCreateDataSource+"."+types.AttributeKeyID][0]))
	ds := h.oracleKeeper.MustGetDataSource(ctx, id)
	h.emitSetDataSource(id, ds, txHash)
	detail["id"] = id
}

// handleMsgCreateOracleScript implements emitter handler for MsgCreateOracleScript.
func (h *Hook) handleMsgCreateOracleScript(
	ctx sdk.Context, txHash []byte, evMap common.EvMap, detail common.JsDict,
) {
	id := types.OracleScriptID(common.Atoi(evMap[types.EventTypeCreateOracleScript+"."+types.AttributeKeyID][0]))
	os := h.oracleKeeper.MustGetOracleScript(ctx, id)
	h.emitSetOracleScript(id, os, txHash)
	detail["id"] = id
}

// handleMsgEditDataSource implements emitter handler for MsgEditDataSource.
func (h *Hook) handleMsgEditDataSource(
	ctx sdk.Context, txHash []byte, msg *types.MsgEditDataSource,
) {
	id := msg.DataSourceID
	ds := h.oracleKeeper.MustGetDataSource(ctx, id)
	h.emitSetDataSource(id, ds, txHash)
}

// handleMsgEditOracleScript implements emitter handler for MsgEditOracleScript.
func (h *Hook) handleMsgEditOracleScript(
	ctx sdk.Context, txHash []byte, msg *types.MsgEditOracleScript,
) {
	id := msg.OracleScriptID
	os := h.oracleKeeper.MustGetOracleScript(ctx, id)
	h.emitSetOracleScript(id, os, txHash)
}

// handleEventRequestExecute implements emitter handler for EventRequestExecute.
func (h *Hook) handleEventRequestExecute(ctx sdk.Context, evMap common.EvMap) {
	if reasons, ok := evMap[types.EventTypeResolve+"."+types.AttributeKeyReason]; ok {
		h.emitUpdateResult(ctx, types.RequestID(common.Atoi(evMap[types.EventTypeResolve+"."+types.AttributeKeyID][0])), reasons[0])
	} else {
		h.emitUpdateResult(ctx, types.RequestID(common.Atoi(evMap[types.EventTypeResolve+"."+types.AttributeKeyID][0])), "")
	}
}

// handleMsgAddReporter implements emitter handler for MsgAddReporter.
func (h *Hook) handleMsgAddReporter(
	ctx sdk.Context, msg *types.MsgAddReporter, detail common.JsDict,
) {
	addr, _ := sdk.ValAddressFromBech32(msg.Validator)
	val, _ := h.stakingKeeper.GetValidator(ctx, addr)
	detail["validator_moniker"] = val.GetMoniker()
	h.AddAccountsInTx(msg.Reporter)
	h.Write("SET_REPORTER", common.JsDict{
		"reporter":  msg.Reporter,
		"validator": msg.Validator,
	})
}

// handleMsgRemoveReporter implements emitter handler for MsgRemoveReporter.
func (h *Hook) handleMsgRemoveReporter(
	ctx sdk.Context, msg *types.MsgRemoveReporter, detail common.JsDict,
) {
	addr, _ := sdk.ValAddressFromBech32(msg.Validator)
	val, _ := h.stakingKeeper.GetValidator(ctx, addr)
	detail["validator_moniker"] = val.GetMoniker()
	h.AddAccountsInTx(msg.Reporter)
	h.Write("REMOVE_REPORTER", common.JsDict{
		"reporter":  msg.Reporter,
		"validator": msg.Validator,
	})
}

// handleMsgActivate implements emitter handler for handleMsgActivate.
func (h *Hook) handleMsgActivate(
	ctx sdk.Context, msg *types.MsgActivate,
) {
	val, _ := sdk.ValAddressFromBech32(msg.Validator)
	h.emitUpdateValidatorStatus(ctx, val)
	h.emitHistoricalValidatorStatus(ctx, val)
}

// handleEventDeactivate implements emitter handler for EventDeactivate.
func (h *Hook) handleEventDeactivate(ctx sdk.Context, evMap common.EvMap) {
	addr, _ := sdk.ValAddressFromBech32(evMap[types.EventTypeDeactivate+"."+types.AttributeKeyValidator][0])
	h.emitUpdateValidatorStatus(ctx, addr)
	h.emitHistoricalValidatorStatus(ctx, addr)
}
