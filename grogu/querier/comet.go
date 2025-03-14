package querier

import (
	"sync/atomic"

	rpcclient "github.com/cometbft/cometbft/rpc/client"

	"github.com/cosmos/cosmos-sdk/client"

	comet "github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
)

type CometQuerier struct {
	queryClients   []comet.ServiceClient
	maxBlockHeight *atomic.Int64
}

func NewCometQuerier(
	clientCtx client.Context,
	clients []rpcclient.RemoteClient,
	maxBlockHeight *atomic.Int64,
) *CometQuerier {
	queryClients := make([]comet.ServiceClient, 0, len(clients))
	for _, cl := range clients {
		queryClients = append(queryClients, comet.NewServiceClient(clientCtx.WithClient(cl)))
	}

	return &CometQuerier{queryClients, maxBlockHeight}
}

func (q *CometQuerier) GetLatestBlock() (*comet.GetLatestBlockResponse, error) {
	fs := make([]QueryFunction[comet.GetLatestBlockRequest, comet.GetLatestBlockResponse], 0, len(q.queryClients))
	for _, queryClient := range q.queryClients {
		fs = append(fs, queryClient.GetLatestBlock)
	}

	in := comet.GetLatestBlockRequest{}
	return getMaxBlockHeightResponse(fs, &in, q.maxBlockHeight)
}
