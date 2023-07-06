package types

import "github.com/bandprotocol/chain/v2/pkg/tss"

// NewResult creates a new Result instance.
func NewResult(
	clientId string,
	oid OracleScriptID,
	calldata []byte,
	askCount, minCount uint64,
	requestId RequestID,
	sid tss.SigningID,
	ansCount uint64,
	requestTime, resolveTime int64,
	resolveStatus ResolveStatus,
	result []byte,
) Result {
	return Result{
		ClientID:       clientId,
		OracleScriptID: oid,
		Calldata:       calldata,
		AskCount:       askCount,
		MinCount:       minCount,
		RequestID:      requestId,
		SigningID:      sid,
		AnsCount:       ansCount,
		RequestTime:    requestTime,
		ResolveTime:    resolveTime,
		ResolveStatus:  resolveStatus,
		Result:         result,
	}
}
