package proof

import (
	context "context"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	gogogrpc "github.com/gogo/protobuf/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/cosmos/cosmos-sdk/client"
)

// RegisterProofService registers the node gRPC service on the provided gRPC router.
func RegisterProofService(clientCtx client.Context, server gogogrpc.Server) {
	RegisterServiceServer(server, NewProofServer(clientCtx))
}

// RegisterGRPCGatewayRoutes mounts the node gRPC service's GRPC-gateway routes
// on the given mux object.
func RegisterGRPCGatewayRoutes(clientConn gogogrpc.ClientConn, mux *runtime.ServeMux) {
	_ = RegisterServiceHandlerClient(context.Background(), mux, NewServiceClient(clientConn))
}

var _ ServiceServer = proofServer{}

type proofServer struct {
	clientCtx client.Context
}

func NewProofServer(clientCtx client.Context) ServiceServer {
	return proofServer{
		clientCtx: clientCtx,
	}
}

func (s proofServer) Proof(ctx context.Context, req *QueryProofRequest) (*QueryProofResponse, error) {
	cliCtx := s.clientCtx
	cliCtx.Height = req.Height
	height := &cliCtx.Height
	if cliCtx.Height == 0 {
		height = nil
	}

	requestID := types.RequestID(req.RequestId)

	commit, err := cliCtx.Client.Commit(context.Background(), height)
	if err != nil {
		return nil, err
	}

	value, iavlEp, multiStoreEp, err := GetProofsByKey(
		cliCtx,
		types.ResultStoreKey(requestID),
		rpcclient.ABCIQueryOptions{Height: commit.Height - 1, Prove: true},
		true,
	)
	if err != nil {
		return nil, err
	}

	signatures, commonVote, err := GetSignaturesAndPrefix(&commit.SignedHeader)
	if err != nil {
		return nil, err
	}
	blockRelay := BlockRelayProof{
		MultiStoreProof:        GetMultiStoreProof(multiStoreEp),
		BlockHeaderMerkleParts: GetBlockHeaderMerkleParts(commit.Header),
		CommonEncodedVotePart:  &commonVote,
		Signatures:             signatures,
	}

	var rs oracletypes.Result
	types.ModuleCdc.MustUnmarshal(value, &rs)

	oracleData := OracleDataProof{
		Result:      &rs,
		Version:     DecodeIAVLLeafPrefix(iavlEp.Leaf.Prefix),
		MerklePaths: GetMerklePaths(iavlEp),
	}

	// Calculate byte for proofbytes
	var relayAndVerifyArguments abi.Arguments
	format := `[{"type":"bytes"},{"type":"bytes"}]`
	err = json.Unmarshal([]byte(format), &relayAndVerifyArguments)
	if err != nil {
		panic(err)
	}

	blockRelayBytes, err := blockRelay.EncodeToEthData()
	if err != nil {
		return nil, err
	}

	oracleDataBytes, err := oracleData.EncodeToEthData(uint64(commit.Height))
	if err != nil {
		return nil, err
	}

	evmProofBytes, err := relayAndVerifyArguments.Pack(blockRelayBytes, oracleDataBytes)
	if err != nil {
		return nil, err
	}

	if cliCtx.Height < 0 {
		return nil, fmt.Errorf("negative height in response")
	}

	return &QueryProofResponse{
		Height: cliCtx.Height,
		Result: &SingleProofResponse{
			Proof: &SingleProof{
				BlockHeight:     uint64(commit.Height),
				OracleDataProof: &oracleData,
				BlockRelayProof: &blockRelay,
			},
			EvmProofBytes: evmProofBytes,
		},
	}, nil
}

func (s proofServer) MultiProof(ctx context.Context, req *QueryMultiProofRequest) (*QueryMultiProofResponse, error) {
	cliCtx := s.clientCtx
	height := &cliCtx.Height
	if cliCtx.Height == 0 {
		height = nil
	}
	requestIDs := req.RequestIds
	if len(requestIDs) == 0 {
		return nil, fmt.Errorf("please provide request ids")
	}

	commit, err := cliCtx.Client.Commit(context.Background(), height)
	if err != nil {
		return nil, err
	}
	signatures, commonVote, err := GetSignaturesAndPrefix(&commit.SignedHeader)
	if err != nil {
		return nil, err
	}

	blockRelay := BlockRelayProof{
		BlockHeaderMerkleParts: GetBlockHeaderMerkleParts(commit.Header),
		CommonEncodedVotePart:  &commonVote,
		Signatures:             signatures,
	}

	oracleDataBytesList := make([][]byte, len(requestIDs))
	oracleDataList := make([]*OracleDataProof, len(requestIDs))

	for idx, intRequestID := range requestIDs {
		requestID := types.RequestID(intRequestID)

		// Extract multiStoreEp in the first iteration only, since multiStoreEp is the same for all requests.
		value, iavlEp, multiStoreEp, err := GetProofsByKey(
			cliCtx,
			types.ResultStoreKey(requestID),
			rpcclient.ABCIQueryOptions{Height: commit.Height - 1, Prove: true},
			idx == 0,
		)
		if err != nil {
			return nil, err
		}

		var rs oracletypes.Result
		types.ModuleCdc.MustUnmarshal(value, &rs)

		oracleData := OracleDataProof{
			Result:      &rs,
			Version:     DecodeIAVLLeafPrefix(iavlEp.Leaf.Prefix),
			MerklePaths: GetMerklePaths(iavlEp),
		}
		oracleDataBytes, err := oracleData.EncodeToEthData(uint64(commit.Height))
		if err != nil {
			return nil, err
		}
		// Append oracle data proof to list
		oracleDataBytesList[idx] = oracleDataBytes
		oracleDataList[idx] = &oracleData

		if idx == 0 {
			blockRelay.MultiStoreProof = GetMultiStoreProof(multiStoreEp)
		}
	}

	blockRelayBytes, err := blockRelay.EncodeToEthData()
	if err != nil {
		return nil, err
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
		return nil, err
	}

	if cliCtx.Height < 0 {
		return nil, fmt.Errorf("negative height in response")
	}

	return &QueryMultiProofResponse{
		Height: cliCtx.Height,
		Result: &MultiProofResponse{
			Proof: &MultiProof{
				BlockHeight:          uint64(commit.Height),
				OracleDataMultiProof: oracleDataList,
				BlockRelayProof:      &blockRelay,
			},
			EvmProofBytes: evmProofBytes,
		},
	}, nil
}

func (s proofServer) RequestCountProof(
	ctx context.Context,
	req *QueryRequestCountProofRequest,
) (*QueryRequestCountProofResponse, error) {
	cliCtx := s.clientCtx
	height := &cliCtx.Height
	if cliCtx.Height == 0 {
		height = nil
	}

	commit, err := cliCtx.Client.Commit(context.Background(), height)
	if err != nil {
		return nil, err
	}

	value, iavlEp, multiStoreEp, err := GetProofsByKey(
		cliCtx,
		types.RequestCountStoreKey,
		rpcclient.ABCIQueryOptions{Height: commit.Height - 1, Prove: true},
		true,
	)
	if err != nil {
		return nil, err
	}

	// Produce block relay proof
	signatures, commonVote, err := GetSignaturesAndPrefix(&commit.SignedHeader)
	if err != nil {
		return nil, err
	}
	blockRelay := BlockRelayProof{
		MultiStoreProof:        GetMultiStoreProof(multiStoreEp),
		BlockHeaderMerkleParts: GetBlockHeaderMerkleParts(commit.Header),
		CommonEncodedVotePart:  &commonVote,
		Signatures:             signatures,
	}

	// Parse requests count
	rs := binary.BigEndian.Uint64(value)

	requestsCountProof := RequestsCountProof{
		Count:       rs,
		Version:     DecodeIAVLLeafPrefix(iavlEp.Leaf.Prefix),
		MerklePaths: GetMerklePaths(iavlEp),
	}

	// Calculate byte for proofbytes
	var relayAndVerifyCountArguments abi.Arguments
	format := `[{"type":"bytes"},{"type":"bytes"}]`
	err = json.Unmarshal([]byte(format), &relayAndVerifyCountArguments)
	if err != nil {
		panic(err)
	}

	blockRelayBytes, err := blockRelay.EncodeToEthData()
	if err != nil {
		return nil, err
	}

	requestsCountBytes, err := requestsCountProof.EncodeToEthData(uint64(commit.Height))
	if err != nil {
		return nil, err
	}

	evmProofBytes, err := relayAndVerifyCountArguments.Pack(blockRelayBytes, requestsCountBytes)
	if err != nil {
		return nil, err
	}

	if cliCtx.Height < 0 {
		return nil, fmt.Errorf("negative height in response")
	}

	return &QueryRequestCountProofResponse{
		Height: cliCtx.Height,
		Result: &CountProofResponse{
			Proof: &CountProof{
				BlockHeight:     uint64(commit.Height),
				CountProof:      &requestsCountProof,
				BlockRelayProof: &blockRelay,
			},
			EvmProofBytes: evmProofBytes,
		},
	}, nil
}
