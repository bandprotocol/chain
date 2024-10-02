package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/globalfee/types"
)

var _ types.QueryServer = &Querier{}

type Querier struct {
	Keeper
}

// Params return parameters of globalfee module
func (q Querier) Params(stdCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(stdCtx)

	return &types.QueryParamsResponse{
		Params: q.GetParams(ctx),
	}, nil
}
