package request

import (
	"encoding/hex"

	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
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

func (h *Hook) getMultiRequestID(requestSearchParams oracletypes.QueryRequestSearchParams, limit int64) []oracletypes.RequestID {
	var requests []Request
	h.dbMap.Select(&requests,
		`select * from request
where oracle_script_id = ? and calldata = ? and min_count = ? and ask_count = ?
order by resolve_time desc limit ?`,
		requestSearchParams.OracleScriptID, requestSearchParams.CallData, requestSearchParams.MinCount, requestSearchParams.AskCount, limit)
	requestIDs := make([]oracletypes.RequestID, len(requests))
	for idx, request := range requests {
		requestIDs[idx] = request.RequestID
	}
	return requestIDs
}
