package querier

import (
	"sync/atomic"

	rpcclient "github.com/cometbft/cometbft/rpc/client"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	feeds "github.com/bandprotocol/chain/v2/x/feeds/types"
)

type FeedQuerier struct {
	queryClients   []feeds.QueryClient
	maxBlockHeight *atomic.Int64
}

func NewFeedQuerier(
	clientCtx client.Context,
	clients []rpcclient.RemoteClient,
	maxBlockHeight *atomic.Int64,
) *FeedQuerier {
	queryClients := make([]feeds.QueryClient, 0, len(clients))
	for _, cl := range clients {
		queryClients = append(queryClients, feeds.NewQueryClient(clientCtx.WithClient(cl)))
	}

	return &FeedQuerier{queryClients, maxBlockHeight}
}

func (q *FeedQuerier) QueryValidValidator(valAddress sdk.ValAddress) (*feeds.QueryValidValidatorResponse, error) {
	fs := make(
		[]QueryFunction[feeds.QueryValidValidatorRequest, feeds.QueryValidValidatorResponse],
		0,
		len(q.queryClients),
	)
	for _, queryClient := range q.queryClients {
		fs = append(fs, queryClient.ValidValidator)
	}

	in := feeds.QueryValidValidatorRequest{
		Validator: valAddress.String(),
	}

	return getMaxBlockHeightResponse(fs, &in, q.maxBlockHeight)
}

func (q *FeedQuerier) QueryIsFeeder(
	validator sdk.ValAddress,
	feeder sdk.Address,
) (*feeds.QueryIsFeederResponse, error) {
	fs := make([]QueryFunction[feeds.QueryIsFeederRequest, feeds.QueryIsFeederResponse], 0, len(q.queryClients))
	for _, queryClient := range q.queryClients {
		fs = append(fs, queryClient.IsFeeder)
	}

	in := feeds.QueryIsFeederRequest{
		FeederAddress:    feeder.String(),
		ValidatorAddress: validator.String(),
	}
	return getMaxBlockHeightResponse(fs, &in, q.maxBlockHeight)
}

func (q *FeedQuerier) QueryValidatorPrices(valAddress sdk.ValAddress) (*feeds.QueryValidatorPricesResponse, error) {
	fs := make(
		[]QueryFunction[feeds.QueryValidatorPricesRequest, feeds.QueryValidatorPricesResponse],
		0,
		len(q.queryClients),
	)
	for _, queryClient := range q.queryClients {
		fs = append(fs, queryClient.ValidatorPrices)
	}

	in := feeds.QueryValidatorPricesRequest{
		Validator: valAddress.String(),
	}
	return getMaxBlockHeightResponse(fs, &in, q.maxBlockHeight)
}

func (q *FeedQuerier) QueryParams() (*feeds.QueryParamsResponse, error) {
	fs := make([]QueryFunction[feeds.QueryParamsRequest, feeds.QueryParamsResponse], 0, len(q.queryClients))
	for _, queryClient := range q.queryClients {
		fs = append(fs, queryClient.Params)
	}

	in := feeds.QueryParamsRequest{}
	return getMaxBlockHeightResponse(fs, &in, q.maxBlockHeight)
}

func (q *FeedQuerier) QueryCurrentFeeds() (*feeds.QueryCurrentFeedsResponse, error) {
	fs := make(
		[]QueryFunction[feeds.QueryCurrentFeedsRequest, feeds.QueryCurrentFeedsResponse],
		0,
		len(q.queryClients),
	)
	for _, queryClient := range q.queryClients {
		fs = append(fs, queryClient.CurrentFeeds)
	}

	in := feeds.QueryCurrentFeedsRequest{}
	return getMaxBlockHeightResponse(fs, &in, q.maxBlockHeight)
}

func (q *FeedQuerier) QueryReferenceSourceConfig() (*feeds.QueryReferenceSourceConfigResponse, error) {
	fs := make(
		[]QueryFunction[feeds.QueryReferenceSourceConfigRequest, feeds.QueryReferenceSourceConfigResponse],
		0,
		len(q.queryClients),
	)
	for _, queryClient := range q.queryClients {
		fs = append(fs, queryClient.ReferenceSourceConfig)
	}

	in := feeds.QueryReferenceSourceConfigRequest{}
	return getMaxBlockHeightResponse(fs, &in, q.maxBlockHeight)
}
