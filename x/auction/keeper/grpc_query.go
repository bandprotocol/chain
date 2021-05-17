package keeper

import (
	"context"
	auctiontypes "github.com/GeoDB-Limited/odin-core/x/auction/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Querier struct {
	Keeper
}

func (q Querier) Params(c context.Context, _ *auctiontypes.QueryParamsRequest) (*auctiontypes.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := q.GetParams(ctx)
	return &auctiontypes.QueryParamsResponse{Params: params}, nil
}

var _ auctiontypes.QueryServer = Querier{}
