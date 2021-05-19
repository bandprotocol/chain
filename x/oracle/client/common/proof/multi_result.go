package proof

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bandprotocol/chain/pkg/obi"
	clientcmn "github.com/bandprotocol/chain/x/oracle/client/common"
	"github.com/bandprotocol/chain/x/oracle/types"
	ics23 "github.com/confio/ics23/go"
	"github.com/cosmos/cosmos-sdk/client"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/ethereum/go-ethereum/accounts/abi"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

type JsonMultiProof struct {
	BlockHeight          uint64            `json:"block_height"`
	OracleDataMultiProof []OracleDataProof `json:"oracle_data_multi_proof"`
	BlockRelayProof      BlockRelayProof   `json:"block_relay_proof"`
}

type MultiProof struct {
	JsonProof     JsonMultiProof   `json:"json_proof"`
	EVMProofBytes tmbytes.HexBytes `json:"evm_proof_bytes"`
}

func GetMutiProofHandlerFn(cliCtx client.Context, route string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestIDs := r.URL.Query()["id"]
		ctx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}
		height := &ctx.Height
		if ctx.Height == 0 {
			height = nil
		}

		commit, err := ctx.Client.Commit(context.Background(), height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		signatures, err := GetSignaturesAndPrefix(&commit.SignedHeader)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		blockRelay := BlockRelayProof{
			BlockHeaderMerkleParts: GetBlockHeaderMerkleParts(commit.Header),
			Signatures:             signatures,
		}

		oracleDataBytesList := make([][]byte, len(requestIDs))
		oracleDataList := make([]OracleDataProof, len(requestIDs))

		isFirstRequest := true
		for idx, requestID := range requestIDs {
			intRequestID, err := strconv.ParseUint(requestID, 10, 64)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			requestID := types.RequestID(intRequestID)
			bz, _, err := ctx.Query(fmt.Sprintf("custom/%s/%s/%d", route, types.QueryRequests, requestID))
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			var qResult types.QueryResult
			if err := json.Unmarshal(bz, &qResult); err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			if qResult.Status != http.StatusOK {
				clientcmn.PostProcessQueryResponse(w, ctx, bz)
				return
			}
			var request types.QueryRequestResult
			if err := ctx.LegacyAmino.UnmarshalJSON(qResult.Result, &request); err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			if request.Result == nil {
				rest.WriteErrorResponse(w, http.StatusNotFound, "Result has not been resolved")
				return
			}

			resp, err := ctx.Client.ABCIQueryWithOptions(
				context.Background(),
				"/store/oracle/key",
				types.ResultStoreKey(requestID),
				rpcclient.ABCIQueryOptions{Height: commit.Height - 1, Prove: true},
			)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}

			proof := resp.Response.GetProofOps()
			if proof == nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, "Proof not found")
				return
			}

			ops := proof.GetOps()
			if ops == nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, "proof ops not found")
				return
			}

			var iavlEp *ics23.ExistenceProof
			var multiStoreEp *ics23.ExistenceProof
			for _, op := range ops {
				switch op.GetType() {
				case storetypes.ProofOpIAVLCommitment:
					proof := &ics23.CommitmentProof{}
					err := proof.Unmarshal(op.Data)
					if err != nil {
						rest.WriteErrorResponse(w, http.StatusInternalServerError,
							fmt.Sprintf("iavl: %s", err.Error()),
						)
						return
					}
					iavlOps := storetypes.NewIavlCommitmentOp(op.Key, proof)
					iavlEp = iavlOps.Proof.GetExist()
					if iavlEp == nil {
						rest.WriteErrorResponse(w, http.StatusNotFound, "Proof has not been ready.")
						return
					}

					resValue := resp.Response.GetValue()

					var rs types.Result
					obi.MustDecode(resValue, &rs)

					oracleData := OracleDataProof{
						Result:      rs,
						Prefix:      iavlEp.Leaf.Prefix,
						MerklePaths: GetMerklePaths(iavlEp),
					}
					oracleDataBytes, err := oracleData.encodeToEthData(uint64(commit.Height))
					if err != nil {
						rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
						return
					}
					// Append oracle data proof to list
					oracleDataBytesList[idx] = oracleDataBytes
					oracleDataList[idx] = oracleData
				case storetypes.ProofOpSimpleMerkleCommitment:
					if isFirstRequest {
						// Only create multi store proof in the first request.
						isFirstRequest = false
						proof := &ics23.CommitmentProof{}
						err := proof.Unmarshal(op.Data)
						if err != nil {
							rest.WriteErrorResponse(w, http.StatusInternalServerError,
								fmt.Sprintf("multiStore: %s", err.Error()),
							)
							return
						}
						multiStoreOps := storetypes.NewSimpleMerkleCommitmentOp(op.Key, proof)
						multiStoreEp = multiStoreOps.Proof.GetExist()
						if multiStoreEp == nil {
							rest.WriteErrorResponse(w, http.StatusNotFound, "Proof has not been ready.")
							return
						}

						blockRelay.MultiStoreProof = GetMultiStoreProof(multiStoreEp)
					}
				default:
					rest.WriteErrorResponse(w, http.StatusInternalServerError,
						fmt.Sprintf("Unknown proof type %s", op.GetType()),
					)
					return
				}
			}
		}

		blockRelayBytes, err := blockRelay.encodeToEthData()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Calculate byte for MultiProofbytes
		var relayAndVerifyArguments abi.Arguments
		format := `[{"type":"bytes"},{"type":"bytes[]"}]`
		err = json.Unmarshal([]byte(format), &relayAndVerifyArguments)
		if err != nil {
			panic(err)
		}

		evmProofBytes, err := relayAndVerifyArguments.Pack(blockRelayBytes, oracleDataBytesList)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, ctx, MultiProof{
			JsonProof: JsonMultiProof{
				BlockHeight:          uint64(commit.Height),
				OracleDataMultiProof: oracleDataList,
				BlockRelayProof:      blockRelay,
			},
			EVMProofBytes: evmProofBytes,
		})
	}
}
