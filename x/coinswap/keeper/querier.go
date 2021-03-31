package keeper

import (
	"github.com/GeoDB-Limited/odin-core/x/coinswap/types"
	commontypes "github.com/GeoDB-Limited/odin-core/x/common/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

func NewQuerier(keeper Keeper, cdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryParams:
			return queryParameters(ctx, keeper, cdc)
		case types.QueryRate:
			return queryRate(ctx, keeper, cdc)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown coinswap query endpoint")
		}
	}
}

func queryParameters(ctx sdk.Context, k Keeper, cdc *codec.LegacyAmino) ([]byte, error) {
	return commontypes.QueryOK(cdc, k.GetParams(ctx))
}

func queryRate(ctx sdk.Context, k Keeper, cdc *codec.LegacyAmino) ([]byte, error) {
	initialRate := k.GetInitialRate(ctx)
	rateMultiplier := k.GetRateMultiplier(ctx)
	return commontypes.QueryOK(cdc, types.QueryRateResult{
		Rate:        initialRate.Mul(rateMultiplier),
		InitialRate: initialRate,
	})
}
