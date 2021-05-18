package proof

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/bandprotocol/chain/x/oracle/types"
	ics23 "github.com/confio/ics23/go"
	"github.com/cosmos/cosmos-sdk/client"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
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

// TODO: Prefix is currently a dynamic-sized bytes, make it fixed
type OracleDataProof struct {
	Result      types.Result     `json:"result"`
	Prefix      tmbytes.HexBytes `json:"prefix"`
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
		o.Prefix,
		parsePaths,
	)
}

type RequestsCountProof struct {
	Count       uint64           `json:"count"`
	Prefix      tmbytes.HexBytes `json:"prefix"`
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
		o.Prefix,
		parsePaths,
	)
}

func getProofsByKey(ctx client.Context, key []byte, queryOptions rpcclient.ABCIQueryOptions) ([]byte, *ics23.ExistenceProof, *ics23.ExistenceProof, error) {
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
				return nil, &ics23.ExistenceProof{}, &ics23.ExistenceProof{}, fmt.Errorf("iavl existence proof not found")
			}
		case storetypes.ProofOpSimpleMerkleCommitment:
			proof := &ics23.CommitmentProof{}
			err := proof.Unmarshal(op.Data)
			if err != nil {
				panic(err)
			}
			multiStoreOps := storetypes.NewSimpleMerkleCommitmentOp(op.Key, proof)
			multiStoreEp = multiStoreOps.Proof.GetExist()
			if multiStoreEp == nil {
				return nil, &ics23.ExistenceProof{}, &ics23.ExistenceProof{}, fmt.Errorf("multistore existence proof not found")
			}
		default:
			return nil, &ics23.ExistenceProof{}, &ics23.ExistenceProof{}, fmt.Errorf("Unknown Proof ops found")
		}
	}

	return resp.Response.GetValue(), iavlEp, multiStoreEp, nil
}
