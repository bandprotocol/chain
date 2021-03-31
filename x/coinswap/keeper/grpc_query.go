package keeper

import (
	"context"
	"github.com/GeoDB-Limited/odin-core/x/coinswap/types"
)

type Querier struct {
	Keeper
}

func (q Querier) Rate(ctx context.Context, request *types.QueryRateRequest) (*types.QueryRateResponse, error) {
	panic("implement me")
}

// todo
//var _ coinswaptypes.QueryServer = Querier{}
