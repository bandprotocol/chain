package types

// NewResult creates a new Result instance.
func NewResult(
	clientID string,
	oid OracleScriptID,
	calldata []byte,
	askCount, minCount uint64,
	requestID RequestID,
	ansCount uint64,
	requestTime, resolveTime int64,
	resolveStatus ResolveStatus,
	result []byte,
) Result {
	return Result{
		ClientID:       clientID,
		OracleScriptID: oid,
		Calldata:       calldata,
		AskCount:       askCount,
		MinCount:       minCount,
		RequestID:      requestID,
		AnsCount:       ansCount,
		RequestTime:    requestTime,
		ResolveTime:    resolveTime,
		ResolveStatus:  resolveStatus,
		Result:         result,
	}
}
