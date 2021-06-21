package proof

import (
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// ResultEthereum is an Ethereum version of Result for solidity ABI-encoding.
type ResultEthereum struct {
	ClientID       string
	OracleScriptID uint64
	Params         []byte
	AskCount       uint64
	MinCount       uint64
	RequestID      uint64
	AnsCount       uint64
	RequestTime    uint64
	ResolveTime    uint64
	ResolveStatus  uint8
	Result         []byte
}

func transformResult(r types.Result) ResultEthereum {
	return ResultEthereum{
		ClientID:       r.ClientID,
		OracleScriptID: uint64(r.OracleScriptID),
		Params:         r.Calldata,
		AskCount:       uint64(r.AskCount),
		MinCount:       uint64(r.MinCount),
		RequestID:      uint64(r.RequestID),
		AnsCount:       uint64(r.AnsCount),
		RequestTime:    uint64(r.RequestTime),
		ResolveTime:    uint64(r.ResolveTime),
		ResolveStatus:  uint8(r.ResolveStatus),
		Result:         r.Result,
	}
}
