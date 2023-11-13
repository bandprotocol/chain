package types

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
)

var (
	fullResult, _ = abi.NewType("tuple", "result", []abi.ArgumentMarshaling{
		{Name: "ClientID", Type: "string"},
		{Name: "OracleScriptID", Type: "uint64"},
		{Name: "Calldata", Type: "bytes"},
		{Name: "AskCount", Type: "uint64"},
		{Name: "MinCount", Type: "uint64"},
		{Name: "RequestID", Type: "uint64"},
		{Name: "AnsCount", Type: "uint64"},
		{Name: "RequestTime", Type: "int64"},
		{Name: "ResolveTime", Type: "int64"},
		{Name: "ResolveStatus", Type: "int32"},
		{Name: "Result", Type: "bytes"},
	})

	fullArgs = abi.Arguments{
		{Type: fullResult, Name: "result"},
	}

	partialResult, _ = abi.NewType("tuple", "result", []abi.ArgumentMarshaling{
		{Name: "Calldata", Type: "bytes"},
		{Name: "OracleScriptID", Type: "uint64"},
		{Name: "RequestID", Type: "uint64"},
		{Name: "MinCount", Type: "uint64"},
		{Name: "ResolveTime", Type: "int64"},
		{Name: "ResolveStatus", Type: "int32"},
		{Name: "Result", Type: "bytes"},
	})

	partialArgs = abi.Arguments{
		{Type: partialResult, Name: "result"},
	}
)

// NewResult creates a new Result instance.
func NewResult(
	clientId string,
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

// PackFullABI serializes a Result struct into a byte slice using the ABI encoding.
func (r Result) PackFullABI() ([]byte, error) {
	abiResult, err := fullArgs.Pack(&r)
	if err != nil {
		return nil, err
	}

	return abiResult, nil
}

// PackPartialABI serializes a Result struct into a byte slice using the ABI encoding with only some fields.
func (r Result) PackPartialABI() ([]byte, error) {
	abiResult, err := partialArgs.Pack(&r)
	if err != nil {
		return nil, err
	}

	return abiResult, nil
}
