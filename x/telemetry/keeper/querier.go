package keeper

import (
	commontypes "github.com/GeoDB-Limited/odin-core/x/common/types"
	telemetrytypes "github.com/GeoDB-Limited/odin-core/x/telemetry/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	abci "github.com/tendermint/tendermint/abci/types"
)

func NewQuerier(keeper Keeper, cdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case telemetrytypes.QueryTopBalances:
			return queryTopBalances(ctx, path[1:], keeper, cdc, req)
		case telemetrytypes.QueryExtendedValidators:
			return queryExtendedValidators(ctx, path[1:], keeper, cdc, req)
		case telemetrytypes.QueryAvgBlockSize:
			return queryAvgBlockSize(ctx, path[1:], keeper, cdc, req)
		case telemetrytypes.QueryAvgBlockTime:
			return queryAvgBlockTime(ctx, path[1:], keeper, cdc, req)
		case telemetrytypes.QueryAvgTxFee:
			return queryAvgTxFee(ctx, path[1:], keeper, cdc, req)
		case telemetrytypes.QueryTxVolume:
			return queryTxVolume(ctx, path[1:], keeper, cdc, req)
		case telemetrytypes.QueryValidatorsBlocks:
			return queryValidatorsBlocks(ctx, path[1:], keeper, cdc, req)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown telemetry query endpoint")
		}
	}
}

func queryTopBalances(
	ctx sdk.Context,
	path []string,
	k Keeper,
	cdc *codec.LegacyAmino,
	req abci.RequestQuery,
) ([]byte, error) {
	if len(path) > 1 {
		return nil, sdkerrors.ErrInvalidRequest
	}
	var params commontypes.QueryPaginationParams
	if err := cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	balances, total := k.GetPaginatedBalances(ctx, path[0], params.Desc, &query.PageRequest{
		Offset: params.Offset,
		Limit:  params.Limit,
	})
	return commontypes.QueryOK(cdc, telemetrytypes.QueryTopBalancesResponse{
		Balances: balances,
		Pagination: &query.PageResponse{
			Total: total,
		},
	})
}

func queryExtendedValidators(
	ctx sdk.Context,
	path []string,
	k Keeper,
	cdc *codec.LegacyAmino,
	req abci.RequestQuery,
) ([]byte, error) {
	if len(path) > 1 {
		return nil, sdkerrors.ErrInvalidRequest
	}
	var params commontypes.QueryPaginationParams
	if err := cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to unmarshal query pagination params")
	}

	var total = 0
	if params.GetCountTotal() {
		total = len(k.stakingQuerier.GetValidators(ctx, k.stakingQuerier.MaxValidators(ctx)))
	}
	validators, err := k.ExtendedValidators(sdk.WrapSDKContext(ctx), &telemetrytypes.QueryExtendedValidatorsRequest{
		Status: path[0],
		Pagination: &query.PageRequest{
			Offset: params.Offset,
			Limit:  params.Limit,
		},
	})
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to query extended validators")
	}

	validators.Pagination.Total = uint64(total)
	return commontypes.QueryOK(cdc, validators)
}

func queryAvgBlockSize(
	_ sdk.Context,
	_ []string,
	k Keeper,
	cdc *codec.LegacyAmino,
	req abci.RequestQuery,
) ([]byte, error) {
	var request telemetrytypes.QueryAvgBlockSizeRequest
	if err := cdc.UnmarshalJSON(req.Data, &request); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	blockSizePerDay, err := k.GetAvgBlockSizePerDay(request.GetStartDate(), request.GetEndDate())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to get average block size per day")
	}
	return commontypes.QueryOK(cdc, telemetrytypes.QueryAvgBlockSizeResponse{
		AvgBlockSizePerDay: blockSizePerDay,
	})
}

func queryAvgBlockTime(
	_ sdk.Context,
	_ []string,
	k Keeper,
	cdc *codec.LegacyAmino,
	req abci.RequestQuery,
) ([]byte, error) {
	var request telemetrytypes.QueryAvgBlockTimeRequest
	if err := cdc.UnmarshalJSON(req.Data, &request); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	blockTimePerDay, err := k.GetAvgBlockTimePerDay(request.GetStartDate(), request.GetEndDate())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to get average block time per day")
	}

	return commontypes.QueryOK(cdc, telemetrytypes.QueryAvgBlockTimeResponse{
		AvgBlockTimePerDay: blockTimePerDay,
	})
}

func queryAvgTxFee(_ sdk.Context, _ []string, k Keeper, cdc *codec.LegacyAmino, req abci.RequestQuery) ([]byte, error) {
	var request telemetrytypes.QueryAvgTxFeeRequest
	if err := cdc.UnmarshalJSON(req.Data, &request); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	avgTxFee, err := k.GetAvgTxFeePerDay(request.GetStartDate(), request.GetEndDate())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to get average tx fee per day")
	}
	return commontypes.QueryOK(cdc, telemetrytypes.QueryAvgTxFeeResponse{
		AvgTxFeePerDay: avgTxFee,
	})
}

func queryTxVolume(_ sdk.Context, _ []string, k Keeper, cdc *codec.LegacyAmino, req abci.RequestQuery) ([]byte, error) {
	var request telemetrytypes.QueryTxVolumeRequest
	if err := cdc.UnmarshalJSON(req.Data, &request); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	txVolume, err := k.GetTxVolumePerDay(request.GetStartDate(), request.GetEndDate())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to get tx volume")
	}

	return commontypes.QueryOK(cdc, telemetrytypes.QueryTxVolumeResponse{
		TxVolumePerDay: txVolume,
	})
}

func queryValidatorsBlocks(
	ctx sdk.Context, _ []string, k Keeper, cdc *codec.LegacyAmino, req abci.RequestQuery,
) ([]byte, error) {
	var request telemetrytypes.QueryValidatorsBlocksRequest
	if err := cdc.UnmarshalJSON(req.Data, &request); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

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

	return commontypes.QueryOK(cdc, telemetrytypes.QueryValidatorsBlocksResponse{
		ValidatorsBlocks: validatorsBlocks,
		Pagination: &query.PageResponse{
			Total: total,
		},
	})
}
