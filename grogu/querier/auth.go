package querier

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type AuthQuerier struct {
	queryClients []auth.QueryClient
}

func NewAuthQuerier(clients []client.Context) *AuthQuerier {
	queryClients := make([]auth.QueryClient, 0, len(clients))
	for _, cl := range clients {
		queryClients = append(queryClients, auth.NewQueryClient(cl))
	}

	return &AuthQuerier{
		queryClients,
	}
}

func (q *AuthQuerier) QueryAccount(address sdk.Address) (*auth.QueryAccountResponse, error) {
	fs := make([]QueryFunction[auth.QueryAccountRequest, auth.QueryAccountResponse], 0, len(q.queryClients))
	for _, queryClient := range q.queryClients {
		fs = append(fs, queryClient.Account)
	}

	in := auth.QueryAccountRequest{Address: address.String()}
	return getMaxBlockHeightResponse(fs, &in)
}
