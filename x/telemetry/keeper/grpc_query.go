package keeper

import (
	"context"
	telemetrytypes "github.com/GeoDB-Limited/odin-core/x/telemetry/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
)

var _ telemetrytypes.QueryServer = Keeper{}

func (k Keeper) TopBalances(
	c context.Context,
	request *telemetrytypes.QueryTopBalancesRequest,
) (*telemetrytypes.QueryTopBalancesResponse, error) {

	ctx := sdk.UnwrapSDKContext(c)
	balances, total := k.GetPaginatedBalances(ctx, request.GetDenom(), request.GetDesc(), request.Pagination)
	return &telemetrytypes.QueryTopBalancesResponse{
		Balances: balances,
		Pagination: &query.PageResponse{
			Total: total,
		},
	}, nil
}

func (k Keeper) AvgBlockSize(
	_ context.Context,
	request *telemetrytypes.QueryAvgBlockSizeRequest,
) (*telemetrytypes.QueryAvgBlockSizeResponse, error) {

	blockSizePerDay, err := k.GetAvgBlockSizePerDay(request.GetStartDate(), request.GetEndDate())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to get average block size per day")
	}

	return &telemetrytypes.QueryAvgBlockSizeResponse{
		AvgBlockSizePerDay: blockSizePerDay,
	}, nil
}

func (k Keeper) AvgBlockTime(
	_ context.Context,
	request *telemetrytypes.QueryAvgBlockTimeRequest,
) (*telemetrytypes.QueryAvgBlockTimeResponse, error) {

	blockTimePerDay, err := k.GetAvgBlockTimePerDay(request.GetStartDate(), request.GetEndDate())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to get average block time per day")
	}

	return &telemetrytypes.QueryAvgBlockTimeResponse{
		AvgBlockTimePerDay: blockTimePerDay,
	}, nil
}

func (k Keeper) AvgTxFee(
	c context.Context,
	request *telemetrytypes.QueryAvgTxFeeRequest,
) (*telemetrytypes.QueryAvgTxFeeResponse, error) {

	avgTxFee, err := k.GetAvgTxFeePerDay(request.GetStartDate(), request.GetEndDate())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to get average tx fee per day")
	}

	return &telemetrytypes.QueryAvgTxFeeResponse{
		AvgTxFeePerDay: avgTxFee,
	}, nil
}

func (k Keeper) TxVolume(
	c context.Context,
	request *telemetrytypes.QueryTxVolumeRequest,
) (*telemetrytypes.QueryTxVolumeResponse, error) {

	txVolume, err := k.GetTxVolumePerDay(request.GetStartDate(), request.GetEndDate())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to get tx volume")
	}

	return &telemetrytypes.QueryTxVolumeResponse{
		TxVolumePerDay: txVolume,
	}, nil
}

func (k Keeper) ValidatorsBlocks(
	c context.Context,
	request *telemetrytypes.QueryValidatorsBlocksRequest,
) (*telemetrytypes.QueryValidatorsBlocksResponse, error) {

	ctx := sdk.UnwrapSDKContext(c)
	validatorsBlocks, total, err := k.GetValidatorsBlocks(
		ctx,
		request.GetStartDate(),
		request.GetEndDate(),
		request.GetDesc(),
		request.GetPagination(),
	)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to get validators blocks")
	}

	return &telemetrytypes.QueryValidatorsBlocksResponse{
		ValidatorsBlocks: validatorsBlocks,
		Pagination: &query.PageResponse{
			Total: total,
		},
	}, nil
}

func (k Keeper) ExtendedValidators(
	c context.Context,
	request *telemetrytypes.QueryExtendedValidatorsRequest,
) (*telemetrytypes.QueryExtendedValidatorsResponse, error) {

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
