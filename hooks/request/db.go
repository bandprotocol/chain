package request

import (
	"encoding/hex"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	sq "github.com/Masterminds/squirrel"
)

const (
	requestsTable = "requests"
)

type Request struct {
	RequestID      oracletypes.RequestID      `db:"request_id, primarykey" json:"request_id"`
	OracleScriptID oracletypes.OracleScriptID `db:"oracle_script_id" json:"oracle_script_id"`
	Calldata       string                     `db:"calldata" json:"calldata"`
	MinCount       uint64                     `db:"min_count" json:"min_count"`
	AskCount       uint64                     `db:"ask_count" json:"ask_count"`
	ResolveTime    int64                      `db:"resolve_time" json:"resolve_time"`
}

func (h *Hook) insertRequest(requestID oracletypes.RequestID, oracleScriptID oracletypes.OracleScriptID, calldata []byte, askCount uint64, minCount uint64, resolveTime int64) {
	err := h.trans.Insert(&Request{
		RequestID:      requestID,
		OracleScriptID: oracleScriptID,
		Calldata:       hex.EncodeToString(calldata),
		MinCount:       minCount,
		AskCount:       askCount,
		ResolveTime:    resolveTime,
	})
	if err != nil {
		panic(err)
	}
}

func (h *Hook) getMultiRequestID(requestSearchRequest oracletypes.QueryRequestSearchRequest, limit int64) oracletypes.QueryRequestIDs {

	requestsSql := sq.Select("*").From(requestsTable)
	conditionsSql := make([]sq.Sqlizer, 0, 4)
	if requestSearchRequest.OracleScriptId != 0 {
		conditionsSql = append(conditionsSql, sq.Eq{"oracle_script_id": requestSearchRequest.OracleScriptId})
	}
	if requestSearchRequest.Calldata != nil {
		conditionsSql = append(conditionsSql, sq.Eq{"calldata": requestSearchRequest.Calldata})
	}
	if requestSearchRequest.MinCount != 0 {
		conditionsSql = append(conditionsSql, sq.Eq{"min_count": requestSearchRequest.MinCount})
	}
	if requestSearchRequest.AskCount != 0 {
		conditionsSql = append(conditionsSql, sq.Eq{"ask_count": requestSearchRequest.AskCount})
	}
	requestsSql = requestsSql.Where(sq.And(conditionsSql)).OrderBy("resolve_time").Limit(uint64(limit))
	rawRequestsSql, args, err := requestsSql.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		panic(err)
	}

	var requests []Request
	_, err = h.dbMap.Select(&requests, rawRequestsSql, args...)
	if err != nil {
		panic(err)
	}

	containerIDs := oracletypes.QueryRequestIDs{
		RequestIds: make([]int64, len(requests)),
	}
	for idx, request := range requests {
		containerIDs.RequestIds[idx] = int64(request.RequestID)
	}

	return containerIDs
}
