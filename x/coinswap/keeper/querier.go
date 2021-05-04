package keeper

import (
	coinswaptypes "github.com/GeoDB-Limited/odin-core/x/coinswap/types"
	commontypes "github.com/GeoDB-Limited/odin-core/x/common/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewQuerier(keeper Keeper, cdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case coinswaptypes.QueryParams:
			return queryParameters(ctx, keeper, cdc)
		case coinswaptypes.QueryRate:
			return queryRate(ctx, req, keeper, cdc)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown coinswap query endpoint")
		}
	}
}

func queryParameters(ctx sdk.Context, k Keeper, cdc *codec.LegacyAmino) ([]byte, error) {
	return commontypes.QueryOK(cdc, k.GetParams(ctx))
}

func queryRate(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, cdc *codec.LegacyAmino) ([]byte, error) {
	var parsedRequest coinswaptypes.QueryRateRequest
	err := cdc.UnmarshalJSON(req.Data, &parsedRequest)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	initialRate := keeper.GetInitialRate(ctx)
	rate, err := keeper.GetRate(ctx, parsedRequest.From, parsedRequest.To)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return commontypes.QueryOK(cdc, coinswaptypes.QueryRateResponse{
		Rate:        rate,
		InitialRate: initialRate,
	})
}
