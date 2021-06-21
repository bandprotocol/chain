package proof

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/gorilla/mux"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

type SingleProof struct {
	BlockHeight     uint64          `json:"block_height"`
	OracleDataProof OracleDataProof `json:"oracle_data_proof"`
	BlockRelayProof BlockRelayProof `json:"block_relay_proof"`
}

type SingleProofResponse struct {
	Proof         SingleProof      `json:"proof"`
	EVMProofBytes tmbytes.HexBytes `json:"evm_proof_bytes"`
}

func GetProofHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}
		height := &ctx.Height
		if ctx.Height == 0 {
			height = nil
		}

		// Parse Request ID
		vars := mux.Vars(r)
		intRequestID, err := strconv.ParseUint(vars[RequestIDTag], 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		requestID := types.RequestID(intRequestID)

		commit, err := ctx.Client.Commit(context.Background(), height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		value, iavlEp, multiStoreEp, err := getProofsByKey(
			ctx,
			types.ResultStoreKey(requestID),
			rpcclient.ABCIQueryOptions{Height: commit.Height - 1, Prove: true},
			true,
		)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		signatures, err := GetSignaturesAndPrefix(&commit.SignedHeader)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		blockRelay := BlockRelayProof{
			MultiStoreProof:        GetMultiStoreProof(multiStoreEp),
			BlockHeaderMerkleParts: GetBlockHeaderMerkleParts(commit.Header),
			Signatures:             signatures,
		}

		var rs types.Result
		types.ModuleCdc.MustUnmarshalBinaryBare(value, &rs)

		oracleData := OracleDataProof{
			Result:      rs,
			Version:     decodeIAVLLeafPrefix(iavlEp.Leaf.Prefix),
			MerklePaths: GetMerklePaths(iavlEp),
		}

		// Calculate byte for proofbytes
		var relayAndVerifyArguments abi.Arguments
		format := `[{"type":"bytes"},{"type":"bytes"}]`
		err = json.Unmarshal([]byte(format), &relayAndVerifyArguments)
		if err != nil {
			panic(err)
		}

		blockRelayBytes, err := blockRelay.encodeToEthData()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		oracleDataBytes, err := oracleData.encodeToEthData(uint64(commit.Height))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		evmProofBytes, err := relayAndVerifyArguments.Pack(blockRelayBytes, oracleDataBytes)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, ctx, SingleProofResponse{
			Proof: SingleProof{
				BlockHeight:     uint64(commit.Height),
				OracleDataProof: oracleData,
				BlockRelayProof: blockRelay,
			},
			EVMProofBytes: evmProofBytes,
		})
	}
}
