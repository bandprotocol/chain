package querier

import (
	"sync/atomic"

	"github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type AuthQuerier struct {
	queryClients   []auth.QueryClient
	maxBlockHeight *atomic.Int64
}

func NewAuthQuerier(
	clientContext client.Context,
	clients []*http.HTTP,
	maxBlockHeight *atomic.Int64,
) *AuthQuerier {
	queryClients := make([]auth.QueryClient, 0, len(clients))
	for _, cl := range clients {
		queryClients = append(queryClients, auth.NewQueryClient(clientContext.WithClient(cl)))
	}

	return &AuthQuerier{
		queryClients,
		maxBlockHeight,
	}
}

func (q *AuthQuerier) QueryAccount(address sdk.Address) (*auth.QueryAccountResponse, error) {
	fs := make([]QueryFunction[auth.QueryAccountRequest, auth.QueryAccountResponse], 0, len(q.queryClients))
	for _, queryClient := range q.queryClients {
		fs = append(fs, queryClient.Account)
	}

	in := auth.QueryAccountRequest{Address: address.String()}
	return getMaxBlockHeightResponse(fs, &in, q.maxBlockHeight)
}