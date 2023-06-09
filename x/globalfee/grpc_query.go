package globalfee

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/globalfee/types"
)

var _ types.QueryServer = &Querier{}

// ParamSource is a read only subset of paramtypes.Subspace
type ParamSource interface {
	Get(ctx sdk.Context, key []byte, ptr interface{})
	Has(ctx sdk.Context, key []byte) bool
}

type Querier struct {
	paramSource ParamSource
}

func NewGrpcQuerier(paramSource ParamSource) Querier {
	return Querier{paramSource: paramSource}
}

// MinimumGasPrices return minimum gas prices
func (q Querier) MinimumGasPrices(
	stdCtx context.Context,
	_ *types.QueryMinimumGasPricesRequest,
) (*types.QueryMinimumGasPricesResponse, error) {
	var minGasPrices sdk.DecCoins
	ctx := sdk.UnwrapSDKContext(stdCtx)
	if q.paramSource.Has(ctx, types.ParamStoreKeyMinGasPrices) {
		q.paramSource.Get(ctx, types.ParamStoreKeyMinGasPrices, &minGasPrices)
	}
	return &types.QueryMinimumGasPricesResponse{
		MinimumGasPrices: minGasPrices,
	}, nil
}
