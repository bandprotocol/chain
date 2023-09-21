package proof

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"

	rpcclient "github.com/cometbft/cometbft/rpc/client"
	"github.com/cosmos/cosmos-sdk/client"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
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

// to check proofServer implements ServiceServer
var _ ServiceServer = proofServer{}

// proofServer implements ServiceServer
type proofServer struct {
	clientCtx client.Context
}

// NewProofServer returns new proofServer from provided client.Context
func NewProofServer(clientCtx client.Context) ServiceServer {
	return proofServer{
		clientCtx: clientCtx,
	}
}

// Proof returns a proof from provided request ID and block height
func (s proofServer) Proof(ctx context.Context, req *QueryProofRequest) (*QueryProofResponse, error) {
	cliCtx := s.clientCtx
	// Set the height in the client context to the requested height
	cliCtx.Height = req.Height
	height := &cliCtx.Height

	// If the height is 0, set the pointer to nil
	if cliCtx.Height == 0 {
		height = nil
	}

	// Convert the request ID to the appropriate type
	requestID := types.RequestID(req.RequestId)

	// Get the commit at the specified height from the client
	commit, err := cliCtx.Client.Commit(context.Background(), height)
	if err != nil {
		return nil, err
	}

	// Get the proofs for the requested id and height
	value, iavlEp, multiStoreEp, err := getProofsByKey(
		cliCtx,
		types.ResultStoreKey(requestID),
		rpcclient.ABCIQueryOptions{Height: commit.Height - 1, Prove: true},
		true,
	)
	if err != nil {
		return nil, err
	}

	// Get the signatures and common vote prefix from the commit header
	signatures, commonVote, err := GetSignaturesAndPrefix(&commit.SignedHeader)
	if err != nil {
		return nil, err
	}

	// Create a BlockRelayProof object with the relevant information
	blockRelay := BlockRelayProof{
		MultiStoreProof:        GetMultiStoreProof(multiStoreEp),
		BlockHeaderMerkleParts: GetBlockHeaderMerkleParts(commit.Header),
		CommonEncodedVotePart:  commonVote,
		Signatures:             signatures,
	}

	// Unmarshal the value into a Result object
	var rs oracletypes.Result
	types.ModuleCdc.MustUnmarshal(value, &rs)

	// Create an OracleDataProof object with the relevant information
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
		return nil, err
	}
	oracleDataBytes, err := oracleData.encodeToEthData(uint64(commit.Height))
	if err != nil {
		return nil, err
	}

	// Pack the encoded block relay bytes and oracle data bytes into a evm proof bytes
	evmProofBytes, err := relayAndVerifyArguments.Pack(blockRelayBytes, oracleDataBytes)
	if err != nil {
		return nil, err
	}

	// If the height is negative, return an error
	if cliCtx.Height < 0 {
		return nil, fmt.Errorf("negative height in response")
	}

	// Return a QueryProofResponse object with the relevant information
	return &QueryProofResponse{
		Height: cliCtx.Height,
		Result: SingleProofResponse{
			Proof: SingleProof{
				BlockHeight:     uint64(commit.Height),
				OracleDataProof: oracleData,
				BlockRelayProof: blockRelay,
			},
			EvmProofBytes: evmProofBytes,
		},
	}, nil
}

// MultiProof returns a proof for multiple request IDs
func (s proofServer) MultiProof(ctx context.Context, req *QueryMultiProofRequest) (*QueryMultiProofResponse, error) {
	// Get the client context from the server context
	cliCtx := s.clientCtx
	height := &cliCtx.Height
	// If the height is 0, set the pointer to nil
	if cliCtx.Height == 0 {
		height = nil
	}
	// Get the request IDs from the request object
	requestIDs := req.RequestIds
	// If there are no request IDs, return an error
	if len(requestIDs) == 0 {
		return nil, fmt.Errorf("please provide request ids")
	}

	// Get the commit at the specified height from the ABCI client
	commit, err := cliCtx.Client.Commit(context.Background(), height)
	if err != nil {
		return nil, err
	}

	// Get the signatures and common vote from the commit header
	signatures, commonVote, err := GetSignaturesAndPrefix(&commit.SignedHeader)
	if err != nil {
		return nil, err
	}

	// Create a BlockRelayProof object with the relevant information
	blockRelay := BlockRelayProof{
		BlockHeaderMerkleParts: GetBlockHeaderMerkleParts(commit.Header),
		CommonEncodedVotePart:  commonVote,
		Signatures:             signatures,
	}

	// Create lists to store the oracle data proof objects and encoded bytes for each request ID
	oracleDataBytesList := make([][]byte, len(requestIDs))
	oracleDataList := make([]OracleDataProof, len(requestIDs))

	// Loop through each request ID and get the relevant proofs
	for idx, intRequestID := range requestIDs {
		requestID := types.RequestID(intRequestID)

		value, iavlEp, multiStoreEp, err := getProofsByKey(
			cliCtx,
			types.ResultStoreKey(requestID),
			rpcclient.ABCIQueryOptions{Height: commit.Height - 1, Prove: true},
			idx == 0,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal the value into a Result object
		var rs oracletypes.Result
		types.ModuleCdc.MustUnmarshal(value, &rs)

		// Create an OracleDataProof object with the relevant information
		oracleData := OracleDataProof{
			Result:      rs,
			Version:     decodeIAVLLeafPrefix(iavlEp.Leaf.Prefix),
			MerklePaths: GetMerklePaths(iavlEp),
		}
		// Encode the oracle data proof into Ethereum-compatible format
		oracleDataBytes, err := oracleData.encodeToEthData(uint64(commit.Height))
		if err != nil {
			return nil, err
		}
		// Append the encoded oracle data proof to the list
		oracleDataBytesList[idx] = oracleDataBytes
		oracleDataList[idx] = oracleData

		// If this is the first iteration, set the multiStoreProof in the blockRelay object
		if idx == 0 {
			blockRelay.MultiStoreProof = GetMultiStoreProof(multiStoreEp)
		}
	}

	// Encode the block relay proof
	blockRelayBytes, err := blockRelay.encodeToEthData()
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

	// Pack the encoded block relay bytes and oracle data bytes into a evm proof bytes
	evmProofBytes, err := relayAndVerifyArguments.Pack(blockRelayBytes, oracleDataBytesList)
	if err != nil {
		return nil, err
	}

	// If the height is negative, return an error
	if cliCtx.Height < 0 {
		return nil, fmt.Errorf("negative height in response")
	}

	// Return a QueryMultiProofResponse object with the relevant information
	return &QueryMultiProofResponse{
		Height: cliCtx.Height,
		Result: MultiProofResponse{
			Proof: MultiProof{
				BlockHeight:          uint64(commit.Height),
				OracleDataMultiProof: oracleDataList,
				BlockRelayProof:      blockRelay,
			},
			EvmProofBytes: evmProofBytes,
		},
	}, nil
}

// RequestCountProof returns a count proof
func (s proofServer) RequestCountProof(
	ctx context.Context,
	req *QueryRequestCountProofRequest,
) (*QueryRequestCountProofResponse, error) {
	// Get the client context from the server context
	cliCtx := s.clientCtx
	height := &cliCtx.Height
	// If the height is 0, set the pointer to nil
	if cliCtx.Height == 0 {
		height = nil
	}

	// Get the commit at the specified height from the client
	commit, err := cliCtx.Client.Commit(context.Background(), height)
	if err != nil {
		return nil, err
	}

	// Get the proofs for the count from the IAVL tree
	value, iavlEp, multiStoreEp, err := getProofsByKey(
		cliCtx,
		types.RequestCountStoreKey,
		rpcclient.ABCIQueryOptions{Height: commit.Height - 1, Prove: true},
		true,
	)
	if err != nil {
		return nil, err
	}

	// Create a BlockRelayProof object with the relevant information
	signatures, commonVote, err := GetSignaturesAndPrefix(&commit.SignedHeader)
	if err != nil {
		return nil, err
	}
	blockRelay := BlockRelayProof{
		MultiStoreProof:        GetMultiStoreProof(multiStoreEp),
		BlockHeaderMerkleParts: GetBlockHeaderMerkleParts(commit.Header),
		CommonEncodedVotePart:  commonVote,
		Signatures:             signatures,
	}

	// Parse the request count from the binary value
	rs := binary.BigEndian.Uint64(value)

	// Create a RequestsCountProof object with the relevant information
	requestsCountProof := RequestsCountProof{
		Count:       rs,
		Version:     decodeIAVLLeafPrefix(iavlEp.Leaf.Prefix),
		MerklePaths: GetMerklePaths(iavlEp),
	}

	// Calculate byte for proofbytes
	var relayAndVerifyCountArguments abi.Arguments
	format := `[{"type":"bytes"},{"type":"bytes"}]`
	err = json.Unmarshal([]byte(format), &relayAndVerifyCountArguments)
	if err != nil {
		panic(err)
	}

	// Encode the block relay proof and the requests count proof into Ethereum-compatible format
	blockRelayBytes, err := blockRelay.encodeToEthData()
	if err != nil {
		return nil, err
	}

	requestsCountBytes, err := requestsCountProof.encodeToEthData(uint64(commit.Height))
	if err != nil {
		return nil, err
	}

	// Pack the encoded proofs into a single byte array
	evmProofBytes, err := relayAndVerifyCountArguments.Pack(blockRelayBytes, requestsCountBytes)
	if err != nil {
		return nil, err
	}

	// If the client context height is negative, return an error
	if cliCtx.Height < 0 {
		return nil, fmt.Errorf("negative height in response")
	}

	// Return the QueryRequestCountProofResponse object with the relevant information
	return &QueryRequestCountProofResponse{
		Height: cliCtx.Height,
		Result: CountProofResponse{
			Proof: CountProof{
				BlockHeight:     uint64(commit.Height),
				CountProof:      requestsCountProof,
				BlockRelayProof: blockRelay,
			},
			EvmProofBytes: evmProofBytes,
		},
	}, nil
}
