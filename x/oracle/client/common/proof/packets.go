package proof

import (
	"github.com/bandprotocol/chain/x/oracle/types"
)

// ResultEthereum is an Ethereum version of Result for solidity ABI-encoding.
type ResultEthereum struct {
	ClientId       string
	OracleScriptId uint64
	Calldata       []byte
	AskCount       uint64
	MinCount       uint64
	RequestId      uint64
	AnsCount       uint64
	RequestTime    uint64
	ResolveTime    uint64
	ResolveStatus  uint8
	Result         []byte
}

func transformResult(r types.Result) ResultEthereum {
	return ResultEthereum{
		ClientId:       r.ClientID,
		OracleScriptId: uint64(r.OracleScriptID),
		Calldata:       r.Calldata,
		AskCount:       uint64(r.AskCount),
		MinCount:       uint64(r.MinCount),
		RequestId:      uint64(r.RequestID),
		AnsCount:       uint64(r.AnsCount),
		RequestTime:    uint64(r.RequestTime),
		ResolveTime:    uint64(r.ResolveTime),
		ResolveStatus:  uint8(r.ResolveStatus),
		Result:         r.Result,
	}
}
