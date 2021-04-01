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

func (q Querier) Params(c context.Context, request *coinswaptypes.QueryParamsRequest) (*coinswaptypes.QueryParamsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	params := q.GetParams(ctx)
	return &coinswaptypes.QueryParamsResponse{Params: params}, nil
}

func (q Querier) Rate(c context.Context, request *coinswaptypes.QueryRateRequest) (*coinswaptypes.QueryRateResponse, error) {
	panic("implement me")
}

// todo
var _ coinswaptypes.QueryServer = Querier{}
