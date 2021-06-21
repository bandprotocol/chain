package proof

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	ics23 "github.com/confio/ics23/go"
	"github.com/cosmos/cosmos-sdk/client"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

var (
	relayArguments       abi.Arguments
	verifyArguments      abi.Arguments
	verifyCountArguments abi.Arguments
)

const (
	RequestIDTag = "requestID"
)

func init() {
	err := json.Unmarshal(relayFormat, &relayArguments)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(verifyFormat, &verifyArguments)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(verifyCountFormat, &verifyCountArguments)
	if err != nil {
		panic(err)
	}
}

type BlockRelayProof struct {
	MultiStoreProof        MultiStoreProof        `json:"multi_store_proof"`
	BlockHeaderMerkleParts BlockHeaderMerkleParts `json:"block_header_merkle_parts"`
	Signatures             []TMSignature          `json:"signatures"`
}

func (blockRelay *BlockRelayProof) encodeToEthData() ([]byte, error) {
	parseSignatures := make([]TMSignatureEthereum, len(blockRelay.Signatures))
	for i, sig := range blockRelay.Signatures {
		parseSignatures[i] = sig.encodeToEthFormat()
	}
	return relayArguments.Pack(
		blockRelay.MultiStoreProof.encodeToEthFormat(),
		blockRelay.BlockHeaderMerkleParts.encodeToEthFormat(),
		parseSignatures,
	)
}

type OracleDataProof struct {
	Result      types.Result     `json:"result"`
	Version     uint64           `json:"version"`
	MerklePaths []IAVLMerklePath `json:"merkle_paths"`
}

func (o *OracleDataProof) encodeToEthData(blockHeight uint64) ([]byte, error) {
	parsePaths := make([]IAVLMerklePathEthereum, len(o.MerklePaths))
	for i, path := range o.MerklePaths {
		parsePaths[i] = path.encodeToEthFormat()
	}
	return verifyArguments.Pack(
		big.NewInt(int64(blockHeight)),
		transformResult(o.Result),
		big.NewInt(int64(o.Version)),
		parsePaths,
	)
}

type RequestsCountProof struct {
	Count       uint64           `json:"count"`
	Version     uint64           `json:"version"`
	MerklePaths []IAVLMerklePath `json:"merkle_paths"`
}

func (o *RequestsCountProof) encodeToEthData(blockHeight uint64) ([]byte, error) {
	parsePaths := make([]IAVLMerklePathEthereum, len(o.MerklePaths))
	for i, path := range o.MerklePaths {
		parsePaths[i] = path.encodeToEthFormat()
	}
	return verifyCountArguments.Pack(
		big.NewInt(int64(blockHeight)),
		big.NewInt(int64(o.Count)),
		big.NewInt(int64(o.Version)),
		parsePaths,
	)
}

func getProofsByKey(ctx client.Context, key []byte, queryOptions rpcclient.ABCIQueryOptions, getMultiStoreEp bool) ([]byte, *ics23.ExistenceProof, *ics23.ExistenceProof, error) {
	resp, err := ctx.Client.ABCIQueryWithOptions(
		context.Background(),
		"/store/oracle/key",
		key,
		queryOptions,
	)
	if err != nil {
		return nil, &ics23.ExistenceProof{}, &ics23.ExistenceProof{}, err
	}

	proof := resp.Response.GetProofOps()
	if proof == nil {
		return nil, &ics23.ExistenceProof{}, &ics23.ExistenceProof{}, fmt.Errorf("Proof not found")
	}

	ops := proof.GetOps()
	if ops == nil {
		return nil, &ics23.ExistenceProof{}, &ics23.ExistenceProof{}, fmt.Errorf("Proof ops not found")
	}

	// Extract iavl proof and multistore existence proof
	var iavlEp *ics23.ExistenceProof
	var multiStoreEp *ics23.ExistenceProof
	for _, op := range ops {
		switch op.GetType() {
		case storetypes.ProofOpIAVLCommitment:
			proof := &ics23.CommitmentProof{}
			err := proof.Unmarshal(op.Data)
			if err != nil {
				panic(err)
			}
			iavlOps := storetypes.NewIavlCommitmentOp(op.Key, proof)
			iavlEp = iavlOps.Proof.GetExist()
			if iavlEp == nil {
				return nil, &ics23.ExistenceProof{}, &ics23.ExistenceProof{}, fmt.Errorf("IAVL existence proof not found")
			}
		case storetypes.ProofOpSimpleMerkleCommitment:
			if getMultiStoreEp {
				proof := &ics23.CommitmentProof{}
				err := proof.Unmarshal(op.Data)
				if err != nil {
					panic(err)
				}
				multiStoreOps := storetypes.NewSimpleMerkleCommitmentOp(op.Key, proof)
				multiStoreEp = multiStoreOps.Proof.GetExist()
				if multiStoreEp == nil {
					return nil, &ics23.ExistenceProof{}, &ics23.ExistenceProof{}, fmt.Errorf("MultiStore existence proof not found")
				}
			}
		default:
			return nil, &ics23.ExistenceProof{}, &ics23.ExistenceProof{}, fmt.Errorf("Unknown proof ops found")
		}
	}

	return resp.Response.GetValue(), iavlEp, multiStoreEp, nil
}
