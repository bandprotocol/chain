package proof

import (
	"github.com/bandprotocol/chain/x/oracle/types"
)

// // RequestPacketEthereum is an Ethereum version of OracleRequestPacketData for solidity ABI-encoding.
// type RequestPacketEthereum struct {
// 	ClientId       string
// 	OracleScriptId uint64
// 	Params         []byte
// 	AskCount       uint64
// 	MinCount       uint64
// }

// func transformRequestPacket(p types.OracleRequestPacketData) RequestPacketEthereum {
// 	return RequestPacketEthereum{
// 		ClientId:       p.ClientID,
// 		OracleScriptId: uint64(p.OracleScriptID),
// 		Params:         p.Calldata,
// 		AskCount:       uint64(p.AskCount),
// 		MinCount:       uint64(p.MinCount),
// 	}
// }

// // ResponsePacketEthereum is an Ethereum version of OracleResponsePacketData for solidity ABI-encoding.
// type ResponsePacketEthereum struct {
// 	ClientId      string
// 	RequestId     uint64
// 	AnsCount      uint64
// 	RequestTime   uint64
// 	ResolveTime   uint64
// 	ResolveStatus uint8
// 	Result        []byte
// }

// func transformResponsePacket(p types.OracleResponsePacketData) ResponsePacketEthereum {
// 	return ResponsePacketEthereum{
// 		ClientId:      p.ClientID,
// 		RequestId:     uint64(p.RequestID),
// 		AnsCount:      uint64(p.AnsCount),
// 		RequestTime:   uint64(p.RequestTime),
// 		ResolveTime:   uint64(p.ResolveTime),
// 		ResolveStatus: uint8(p.ResolveStatus),
// 		Result:        p.Result,
// 	}
// }

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
