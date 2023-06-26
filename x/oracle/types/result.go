package types

import "github.com/bandprotocol/chain/v2/pkg/tss"

// NewResult creates a new Result instance.
func NewResult(
	clientId string,
	gid tss.GroupID,
	oid OracleScriptID,
	calldata []byte,
	askCount, minCount uint64,
	requestId RequestID,
	ansCount uint64,
	requestTime, resolveTime int64,
	resolveStatus ResolveStatus,
	result []byte,
) Result {
	return Result{
		ClientID:       clientId,
		GroupID:        gid,
		OracleScriptID: oid,
		Calldata:       calldata,
		AskCount:       askCount,
		MinCount:       minCount,
		RequestID:      requestId,
		AnsCount:       ansCount,
		RequestTime:    requestTime,
		ResolveTime:    resolveTime,
		ResolveStatus:  resolveStatus,
		Result:         result,
	}
}
