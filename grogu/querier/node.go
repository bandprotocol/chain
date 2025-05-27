package querier

import (
	"sync/atomic"

	rpcclient "github.com/cometbft/cometbft/rpc/client"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/node"
)

type NodeQuerier struct {
	queryClients   []node.ServiceClient
	maxBlockHeight *atomic.Int64
}

func NewNodeQuerier(
	clientCtx client.Context,
	clients []rpcclient.RemoteClient,
	maxBlockHeight *atomic.Int64,
) *NodeQuerier {
	queryClients := make([]node.ServiceClient, 0, len(clients))
	for _, cl := range clients {
		queryClients = append(queryClients, node.NewServiceClient(clientCtx.WithClient(cl)))
	}

	return &NodeQuerier{queryClients, maxBlockHeight}
}

func (q *NodeQuerier) QueryStatus() (*node.StatusResponse, error) {
	fs := make(
		[]QueryFunction[node.StatusRequest, node.StatusResponse],
		0,
		len(q.queryClients),
	)
	for _, queryClient := range q.queryClients {
		fs = append(fs, queryClient.Status)
	}

	in := node.StatusRequest{}

	return getMaxBlockHeightResponse(fs, &in, q.maxBlockHeight)
}
