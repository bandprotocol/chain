package keeper

import (
	"context"
	telemetrytypes "github.com/GeoDB-Limited/odin-core/x/telemetry/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
)

var _ telemetrytypes.QueryServer = Keeper{}

func (k Keeper) TopBalances(c context.Context, request *telemetrytypes.QueryTopBalancesRequest) (*telemetrytypes.QueryTopBalancesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	balances, total := k.GetPaginatedBalances(ctx, request.GetDenom(), request.GetDesc(), request.Pagination)
	return &telemetrytypes.QueryTopBalancesResponse{
		Balances: balances,
		Pagination: &query.PageResponse{
			Total: total,
		},
	}, nil
}

func (k Keeper) ExtendedValidators(c context.Context, request *telemetrytypes.QueryExtendedValidatorsRequest) (*telemetrytypes.QueryExtendedValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	validatorsResp, err := k.stakingQuerier.Validators(c, ExtendedValidatorsRequestToValidatorsRequest(request))
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to get validators")
	}
	accounts, err := ValidatorsToAccounts(validatorsResp.GetValidators())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to get validators accounts addresses")
	}
	extendedValidatorsResp := ValidatorsResponseToExtendedValidatorsResponse(validatorsResp)
	extendedValidatorsResp.Balances = k.GetBalances(ctx, accounts...)
	return extendedValidatorsResp, nil
}
