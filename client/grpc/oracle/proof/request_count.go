package proof

// import (
// 	"context"
// 	"encoding/binary"
// 	"encoding/json"
// 	"net/http"

// 	"github.com/cosmos/cosmos-sdk/client"
// 	"github.com/cosmos/cosmos-sdk/types/rest"
// 	"github.com/ethereum/go-ethereum/accounts/abi"
// 	tmbytes "github.com/tendermint/tendermint/libs/bytes"
// 	rpcclient "github.com/tendermint/tendermint/rpc/client"

// 	"github.com/bandprotocol/chain/v2/x/oracle/types"
// )

// type CountProof struct {
// 	BlockHeight     uint64             `json:"block_height"`
// 	CountProof      RequestsCountProof `json:"count_proof"`
// 	BlockRelayProof BlockRelayProof    `json:"block_relay_proof"`
// }

// type CountProofResponse struct {
// 	Proof         CountProof       `json:"proof"`
// 	EVMProofBytes tmbytes.HexBytes `json:"evm_proof_bytes"`
// }

// func GetRequestsCountProofHandlerFn(cliCtx client.Context) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		ctx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
// 		if !ok {
// 			return
// 		}
// 		height := &ctx.Height
// 		if ctx.Height == 0 {
// 			height = nil
// 		}

// 		commit, err := ctx.Client.Commit(context.Background(), height)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
// 			return
// 		}

// 		value, iavlEp, multiStoreEp, err := getProofsByKey(
// 			ctx,
// 			types.RequestCountStoreKey,
// 			rpcclient.ABCIQueryOptions{Height: commit.Height - 1, Prove: true},
// 			true,
// 		)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
// 			return
// 		}

// 		// Produce block relay proof
// 		signatures, commonVote, err := GetSignaturesAndPrefix(&commit.SignedHeader)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}
// 		blockRelay := BlockRelayProof{
// 			MultiStoreProof:        GetMultiStoreProof(multiStoreEp),
// 			BlockHeaderMerkleParts: GetBlockHeaderMerkleParts(commit.Header),
// 			CommonEncodedVotePart:  commonVote,
// 			Signatures:             signatures,
// 		}

// 		// Parse requests count
// 		rs := binary.BigEndian.Uint64(value)

// 		requestsCountProof := RequestsCountProof{
// 			Count:       rs,
// 			Version:     decodeIAVLLeafPrefix(iavlEp.Leaf.Prefix),
// 			MerklePaths: GetMerklePaths(iavlEp),
// 		}

// 		// Calculate byte for proofbytes
// 		var relayAndVerifyCountArguments abi.Arguments
// 		format := `[{"type":"bytes"},{"type":"bytes"}]`
// 		err = json.Unmarshal([]byte(format), &relayAndVerifyCountArguments)
// 		if err != nil {
// 			panic(err)
// 		}

// 		blockRelayBytes, err := blockRelay.encodeToEthData()
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}

// 		requestsCountBytes, err := requestsCountProof.encodeToEthData(uint64(commit.Height))
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}

// 		evmProofBytes, err := relayAndVerifyCountArguments.Pack(blockRelayBytes, requestsCountBytes)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}

// 		rest.PostProcessResponse(w, ctx, CountProofResponse{
// 			Proof: CountProof{
// 				BlockHeight:     uint64(commit.Height),
// 				CountProof:      requestsCountProof,
// 				BlockRelayProof: blockRelay,
// 			},
// 			EVMProofBytes: evmProofBytes,
// 		})
// 	}
// }
