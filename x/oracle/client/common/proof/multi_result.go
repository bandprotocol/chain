package proof

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"strconv"

// 	"github.com/bandprotocol/chain/pkg/obi"
// 	clientcmn "github.com/bandprotocol/chain/x/oracle/client/common"
// 	"github.com/bandprotocol/chain/x/oracle/types"
// 	"github.com/cosmos/cosmos-sdk/client/context"
// 	"github.com/cosmos/cosmos-sdk/store/rootmulti"
// 	"github.com/cosmos/cosmos-sdk/types/rest"
// 	"github.com/ethereum/go-ethereum/accounts/abi"
// 	"github.com/tendermint/iavl"
// 	tmbytes "github.com/tendermint/tendermint/libs/bytes"
// 	rpcclient "github.com/tendermint/tendermint/rpc/client"
// )

// type JsonMultiProof struct {
// 	BlockHeight          uint64            `json:"blockHeight"`
// 	OracleDataMultiProof []OracleDataProof `json:"oracleDataMultiProof"`
// 	BlockRelayProof      BlockRelayProof   `json:"blockRelayProof"`
// }

// type MultiProof struct {
// 	JsonProof     JsonMultiProof   `json:"jsonProof"`
// 	EVMProofBytes tmbytes.HexBytes `json:"evmProofBytes"`
// }

// func GetMutiProofHandlerFn(cliCtx context.CLIContext, route string) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		requestIDs := r.URL.Query()["id"]
// 		ctx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
// 		if !ok {
// 			return
// 		}
// 		height := &ctx.Height
// 		if ctx.Height == 0 {
// 			height = nil
// 		}

// 		commit, err := ctx.Client.Commit(height)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}
// 		signatures, err := GetSignaturesAndPrefix(&commit.SignedHeader)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}

// 		blockRelay := BlockRelayProof{
// 			BlockHeaderMerkleParts: GetBlockHeaderMerkleParts(ctx.Codec, commit.Header),
// 			Signatures:             signatures,
// 		}

// 		oracleDataBytesList := make([][]byte, len(requestIDs))
// 		oracleDataList := make([]OracleDataProof, len(requestIDs))

// 		isFirstRequest := true
// 		for idx, requestID := range requestIDs {
// 			intRequestID, err := strconv.ParseUint(requestID, 10, 64)
// 			if err != nil {
// 				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
// 				return
// 			}
// 			requestID := types.RequestID(intRequestID)
// 			bz, _, err := ctx.Query(fmt.Sprintf("custom/%s/%s/%d", route, types.QueryRequests, requestID))
// 			if err != nil {
// 				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 				return
// 			}
// 			var qResult types.QueryResult
// 			if err := json.Unmarshal(bz, &qResult); err != nil {
// 				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 				return
// 			}
// 			if qResult.Status != http.StatusOK {
// 				clientcmn.PostProcessQueryResponse(w, ctx, bz)
// 				return
// 			}
// 			var request types.QueryRequestResult
// 			if err := ctx.Codec.UnmarshalJSON(qResult.Result, &request); err != nil {
// 				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 				return
// 			}
// 			if request.Result == nil {
// 				rest.WriteErrorResponse(w, http.StatusNotFound, "Result has not been resolved")
// 				return
// 			}

// 			resp, err := ctx.Client.ABCIQueryWithOptions(
// 				"/store/oracle/key",
// 				types.ResultStoreKey(requestID),
// 				rpcclient.ABCIQueryOptions{Height: commit.Height - 1, Prove: true},
// 			)
// 			if err != nil {
// 				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 				return
// 			}

// 			proof := resp.Response.GetProof()
// 			if proof == nil {
// 				rest.WriteErrorResponse(w, http.StatusInternalServerError, "Proof not found")
// 				return
// 			}

// 			ops := proof.GetOps()
// 			if ops == nil {
// 				rest.WriteErrorResponse(w, http.StatusInternalServerError, "proof ops not found")
// 				return
// 			}

// 			var iavlProof iavl.ValueOp
// 			var multiStoreProof rootmulti.MultiStoreProofOp
// 			for _, op := range ops {
// 				switch op.GetType() {
// 				case "iavl:v":
// 					err := ctx.Codec.UnmarshalBinaryLengthPrefixed(op.GetData(), &iavlProof)
// 					if err != nil {
// 						rest.WriteErrorResponse(w, http.StatusInternalServerError,
// 							fmt.Sprintf("iavl: %s", err.Error()),
// 						)
// 						return
// 					}
// 					if iavlProof.Proof == nil {
// 						rest.WriteErrorResponse(w, http.StatusNotFound, "Proof has not been ready.")
// 						return
// 					}

// 					eventHeight := iavlProof.Proof.Leaves[0].Version
// 					resValue := resp.Response.GetValue()

// 					type result struct {
// 						Req types.OracleRequestPacketData
// 						Res types.OracleResponsePacketData
// 					}
// 					var rs result
// 					obi.MustDecode(resValue, &rs)

// 					oracleData := OracleDataProof{
// 						RequestPacket:  rs.Req,
// 						ResponsePacket: rs.Res,
// 						Version:        uint64(eventHeight),
// 						MerklePaths:    GetIAVLMerklePaths(&iavlProof),
// 					}
// 					oracleDataBytes, err := oracleData.encodeToEthData(uint64(commit.Height))
// 					if err != nil {
// 						rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 						return
// 					}
// 					// Append oracle data proof to list
// 					oracleDataBytesList[idx] = oracleDataBytes
// 					oracleDataList[idx] = oracleData
// 				case "multistore":
// 					if isFirstRequest {
// 						// Only create multi store proof in the first request.
// 						isFirstRequest = false
// 						mp, err := rootmulti.MultiStoreProofOpDecoder(op)
// 						multiStoreProof = mp.(rootmulti.MultiStoreProofOp)
// 						if err != nil {
// 							rest.WriteErrorResponse(w, http.StatusInternalServerError,
// 								fmt.Sprintf("multiStore: %s", err.Error()),
// 							)
// 							return
// 						}
// 						blockRelay.MultiStoreProof = GetMultiStoreProof(multiStoreProof)
// 					}
// 				case "iavl:a":
// 					rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf(
// 						"Proof of #%d is unavailable please wait on the next block", requestID,
// 					))
// 					return
// 				default:
// 					rest.WriteErrorResponse(w, http.StatusInternalServerError,
// 						fmt.Sprintf("Unknown proof type %s", op.GetType()),
// 					)
// 					return
// 				}
// 			}
// 		}

// 		blockRelayBytes, err := blockRelay.encodeToEthData()
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}

// 		// Calculate byte for MultiProofbytes
// 		var relayAndVerifyArguments abi.Arguments
// 		format := `[{"type":"bytes"},{"type":"bytes[]"}]`
// 		err = json.Unmarshal([]byte(format), &relayAndVerifyArguments)
// 		if err != nil {
// 			panic(err)
// 		}

// 		evmProofBytes, err := relayAndVerifyArguments.Pack(blockRelayBytes, oracleDataBytesList)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}

// 		rest.PostProcessResponse(w, ctx, MultiProof{
// 			JsonProof: JsonMultiProof{
// 				BlockHeight:          uint64(commit.Height),
// 				OracleDataMultiProof: oracleDataList,
// 				BlockRelayProof:      blockRelay,
// 			},
// 			EVMProofBytes: evmProofBytes,
// 		})
// 	}
// }
