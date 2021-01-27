package proof

// import (
// 	"encoding/json"
// 	"fmt"
// 	"math/big"

// 	"github.com/cosmos/cosmos-sdk/client/context"
// 	"github.com/cosmos/cosmos-sdk/store/rootmulti"
// 	"github.com/ethereum/go-ethereum/accounts/abi"
// 	"github.com/tendermint/iavl"
// 	rpcclient "github.com/tendermint/tendermint/rpc/client"

// 	"github.com/bandprotocol/chain/x/oracle/types"
// )

// var (
// 	relayArguments       abi.Arguments
// 	verifyArguments      abi.Arguments
// 	verifyCountArguments abi.Arguments
// )

// const (
// 	RequestIDTag = "requestID"
// )

// func init() {
// 	err := json.Unmarshal(relayFormat, &relayArguments)
// 	if err != nil {
// 		panic(err)
// 	}
// 	err = json.Unmarshal(verifyFormat, &verifyArguments)
// 	if err != nil {
// 		panic(err)
// 	}
// 	err = json.Unmarshal(verifyCountFormat, &verifyCountArguments)
// 	if err != nil {
// 		panic(err)
// 	}
// }

// type BlockRelayProof struct {
// 	MultiStoreProof        MultiStoreProof        `json:"multiStoreProof"`
// 	BlockHeaderMerkleParts BlockHeaderMerkleParts `json:"blockHeaderMerkleParts"`
// 	Signatures             []TMSignature          `json:"signatures"`
// }

// func (blockRelay *BlockRelayProof) encodeToEthData() ([]byte, error) {
// 	parseSignatures := make([]TMSignatureEthereum, len(blockRelay.Signatures))
// 	for i, sig := range blockRelay.Signatures {
// 		parseSignatures[i] = sig.encodeToEthFormat()
// 	}
// 	return relayArguments.Pack(
// 		blockRelay.MultiStoreProof.encodeToEthFormat(),
// 		blockRelay.BlockHeaderMerkleParts.encodeToEthFormat(),
// 		parseSignatures,
// 	)
// }

// type OracleDataProof struct {
// 	RequestPacket  types.OracleRequestPacketData  `json:"requestPacket"`
// 	ResponsePacket types.OracleResponsePacketData `json:"responsePacket"`
// 	Version        uint64                         `json:"version"`
// 	MerklePaths    []IAVLMerklePath               `json:"merklePaths"`
// }

// func (o *OracleDataProof) encodeToEthData(blockHeight uint64) ([]byte, error) {
// 	parsePaths := make([]IAVLMerklePathEthereum, len(o.MerklePaths))
// 	for i, path := range o.MerklePaths {
// 		parsePaths[i] = path.encodeToEthFormat()
// 	}
// 	return verifyArguments.Pack(
// 		big.NewInt(int64(blockHeight)),
// 		transformRequestPacket(o.RequestPacket),
// 		transformResponsePacket(o.ResponsePacket),
// 		big.NewInt(int64(o.Version)),
// 		parsePaths,
// 	)
// }

// type RequestsCountProof struct {
// 	Count       uint64           `json:"count"`
// 	Version     uint64           `json:"version"`
// 	MerklePaths []IAVLMerklePath `json:"merklePaths"`
// }

// func (o *RequestsCountProof) encodeToEthData(blockHeight uint64) ([]byte, error) {
// 	parsePaths := make([]IAVLMerklePathEthereum, len(o.MerklePaths))
// 	for i, path := range o.MerklePaths {
// 		parsePaths[i] = path.encodeToEthFormat()
// 	}
// 	return verifyCountArguments.Pack(
// 		big.NewInt(int64(blockHeight)),
// 		big.NewInt(int64(o.Count)),
// 		big.NewInt(int64(o.Version)),
// 		parsePaths,
// 	)
// }

// func getProofsByKey(ctx context.CLIContext, key []byte, queryOptions rpcclient.ABCIQueryOptions) ([]byte, iavl.ValueOp, rootmulti.MultiStoreProofOp, error) {
// 	resp, err := ctx.Client.ABCIQueryWithOptions(
// 		"/store/oracle/key",
// 		key,
// 		queryOptions,
// 	)
// 	if err != nil {
// 		return nil, iavl.ValueOp{}, rootmulti.MultiStoreProofOp{}, err
// 	}

// 	proof := resp.Response.GetProof()
// 	if proof == nil {
// 		return nil, iavl.ValueOp{}, rootmulti.MultiStoreProofOp{}, fmt.Errorf("Proof not found")
// 	}

// 	ops := proof.GetOps()
// 	if ops == nil {
// 		return nil, iavl.ValueOp{}, rootmulti.MultiStoreProofOp{}, fmt.Errorf("Proof ops not found")
// 	}

// 	// Extract iavl proof and multi store proof
// 	var iavlProof iavl.ValueOp
// 	var multiStoreProof rootmulti.MultiStoreProofOp
// 	for _, op := range ops {

// 		opType := op.GetType()

// 		if opType == "iavl:v" {
// 			err := ctx.Codec.UnmarshalBinaryLengthPrefixed(op.GetData(), &iavlProof)
// 			if err != nil {
// 				return nil, iavl.ValueOp{}, rootmulti.MultiStoreProofOp{}, fmt.Errorf("iavl: %s", err.Error())
// 			}
// 		} else if opType == "multistore" {
// 			mp, err := rootmulti.MultiStoreProofOpDecoder(op)

// 			multiStoreProof = mp.(rootmulti.MultiStoreProofOp)
// 			if err != nil {
// 				return nil, iavl.ValueOp{}, rootmulti.MultiStoreProofOp{}, fmt.Errorf("multiStore: %s", err.Error())
// 			}
// 		}
// 	}

// 	return resp.Response.GetValue(), iavlProof, multiStoreProof, nil
// }
