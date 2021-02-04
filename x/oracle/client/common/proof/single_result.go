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
// 	"github.com/cosmos/cosmos-sdk/types/rest"
// 	"github.com/ethereum/go-ethereum/accounts/abi"
// 	"github.com/gorilla/mux"
// 	tmbytes "github.com/tendermint/tendermint/libs/bytes"
// 	rpcclient "github.com/tendermint/tendermint/rpc/client"
// )

// type JsonProof struct {
// 	BlockHeight     uint64          `json:"blockHeight"`
// 	OracleDataProof OracleDataProof `json:"oracleDataProof"`
// 	BlockRelayProof BlockRelayProof `json:"blockRelayProof"`
// }

// type Proof struct {
// 	JsonProof     JsonProof        `json:"jsonProof"`
// 	EVMProofBytes tmbytes.HexBytes `json:"evmProofBytes"`
// }

// func GetProofHandlerFn(cliCtx context.CLIContext, route string) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		ctx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
// 		if !ok {
// 			return
// 		}
// 		height := &ctx.Height
// 		if ctx.Height == 0 {
// 			height = nil
// 		}

// 		// Parse Request ID
// 		vars := mux.Vars(r)
// 		intRequestID, err := strconv.ParseUint(vars[RequestIDTag], 10, 64)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
// 			return
// 		}
// 		requestID := types.RequestID(intRequestID)

// 		// Get Request and proof
// 		bz, _, err := ctx.Query(fmt.Sprintf("custom/%s/%s/%d", route, types.QueryRequests, requestID))
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}
// 		var qResult types.QueryResult
// 		if err := json.Unmarshal(bz, &qResult); err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}
// 		if qResult.Status != http.StatusOK {
// 			clientcmn.PostProcessQueryResponse(w, ctx, bz)
// 			return
// 		}
// 		var request types.QueryRequestResult
// 		if err := ctx.Codec.UnmarshalJSON(qResult.Result, &request); err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}
// 		if request.Result == nil {
// 			rest.WriteErrorResponse(w, http.StatusNotFound, "Result has not been resolved")
// 			return
// 		}

// 		commit, err := ctx.Client.Commit(height)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}

// 		value, iavlProof, multiStoreProof, err := getProofsByKey(
// 			ctx,
// 			types.ResultStoreKey(requestID),
// 			rpcclient.ABCIQueryOptions{Height: commit.Height - 1, Prove: true},
// 		)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}

// 		if iavlProof.Proof == nil {
// 			rest.WriteErrorResponse(w, http.StatusNotFound, "Proof has not been ready.")
// 			return
// 		}
// 		eventHeight := iavlProof.Proof.Leaves[0].Version
// 		signatures, err := GetSignaturesAndPrefix(&commit.SignedHeader)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}
// 		blockRelay := BlockRelayProof{
// 			MultiStoreProof:        GetMultiStoreProof(multiStoreProof),
// 			BlockHeaderMerkleParts: GetBlockHeaderMerkleParts(ctx.Codec, commit.Header),
// 			Signatures:             signatures,
// 		}

// 		type result struct {
// 			Req types.OracleRequestPacketData
// 			Res types.OracleResponsePacketData
// 		}
// 		var rs result
// 		obi.MustDecode(value, &rs)

// 		oracleData := OracleDataProof{
// 			RequestPacket:  rs.Req,
// 			ResponsePacket: rs.Res,
// 			Version:        uint64(eventHeight),
// 			MerklePaths:    GetIAVLMerklePaths(&iavlProof),
// 		}

// 		// Calculate byte for proofbytes
// 		var relayAndVerifyArguments abi.Arguments
// 		format := `[{"type":"bytes"},{"type":"bytes"}]`
// 		err = json.Unmarshal([]byte(format), &relayAndVerifyArguments)
// 		if err != nil {
// 			panic(err)
// 		}

// 		blockRelayBytes, err := blockRelay.encodeToEthData()
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}

// 		oracleDataBytes, err := oracleData.encodeToEthData(uint64(commit.Height))
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}

// 		evmProofBytes, err := relayAndVerifyArguments.Pack(blockRelayBytes, oracleDataBytes)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}

// 		rest.PostProcessResponse(w, ctx, Proof{
// 			JsonProof: JsonProof{
// 				BlockHeight:     uint64(commit.Height),
// 				OracleDataProof: oracleData,
// 				BlockRelayProof: blockRelay,
// 			},
// 			EVMProofBytes: evmProofBytes,
// 		})
// 	}
// }
