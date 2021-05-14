package yoda

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/bandprotocol/chain/x/oracle/types"
)

func getPendingRequests(grpcURI string, req *types.QueryPendingRequestsRequest) (*types.QueryPendingRequestsResponse, error) {
	grpcConn, err := grpc.Dial(grpcURI, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("unable to dial gRPC connection: %w", err)
	}
	defer grpcConn.Close()

	client := types.NewQueryClient(grpcConn)
	res, err := client.PendingRequests(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("unable to query pending requests: %w", err)
	}

	return res, nil
}
