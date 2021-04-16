package keeper

import (
	"context"
	coinswaptypes "github.com/GeoDB-Limited/odin-core/x/coinswap/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Querier struct {
	Keeper
}

func (q Querier) Params(c context.Context, _ *coinswaptypes.QueryParamsRequest) (*coinswaptypes.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := q.GetParams(ctx)
	return &coinswaptypes.QueryParamsResponse{Params: params}, nil
}

func (q Querier) Rate(c context.Context, request *coinswaptypes.QueryRateRequest) (*coinswaptypes.QueryRateResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	initialRate := q.GetInitialRate(ctx)
	rateMultiplier, err := q.GetRate(ctx, request.From, request.To)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &coinswaptypes.QueryRateResponse{
		Rate:        initialRate.Mul(rateMultiplier),
		InitialRate: initialRate,
	}, nil
}

var _ coinswaptypes.QueryServer = Querier{}
