package oraclekeeper

import (
	commontypes "github.com/GeoDB-Limited/odin-core/x/common/types"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"strconv"
)

// NewQuerier is the module level router for state queries.
func NewQuerier(keeper Keeper, cdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case oracletypes.QueryParams:
			return queryParameters(ctx, keeper, req, cdc)
		case oracletypes.QueryCounts:
			return queryCounts(ctx, keeper, req, cdc)
		case oracletypes.QueryData:
			return queryData(ctx, path[1:], keeper, req, cdc)
		case oracletypes.QueryDataSources:
			return queryDataSources(ctx, path[1:], keeper, req, cdc)
		case oracletypes.QueryOracleScripts:
			return queryOracleScripts(ctx, path[1:], keeper, req, cdc)
		case oracletypes.QueryRequests:
			return queryRequests(ctx, path[1:], keeper, req, cdc)
		case oracletypes.QueryRequestReports:
			return queryRequestReports(ctx, path[1:], keeper, req, cdc)
		case oracletypes.QueryValidatorStatus:
			return queryValidatorStatus(ctx, path[1:], keeper, req, cdc)
		case oracletypes.QueryReporters:
			return queryReporters(ctx, path[1:], keeper, req, cdc)
		case oracletypes.QueryActiveValidators:
			return queryActiveValidators(ctx, keeper, req, cdc)
		case oracletypes.QueryPendingRequests:
			return queryPendingRequests(ctx, path[1:], keeper, req, cdc)
		case oracletypes.QueryDataProvidersPool:
			return queryDataProvidersPool(ctx, keeper, req, cdc)
		case oracletypes.QueryDataProviderReward:
			return queryDataProvidersRewardPerByte(ctx, keeper, req, cdc)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown oracle query endpoint")
		}
	}
}

func queryParameters(ctx sdk.Context, k Keeper, _ abci.RequestQuery, cdc *codec.LegacyAmino) ([]byte, error) {
	return commontypes.QueryOK(cdc, k.GetParams(ctx))
}

func queryCounts(ctx sdk.Context, k Keeper, req abci.RequestQuery, cdc *codec.LegacyAmino) ([]byte, error) {
	return commontypes.QueryOK(cdc, oracletypes.QueryCountsResponse{
		DataSourceCount:   k.GetDataSourceCount(ctx),
		OracleScriptCount: k.GetOracleScriptCount(ctx),
		RequestCount:      k.GetRequestCount(ctx),
	})
}

func queryData(ctx sdk.Context, path []string, k Keeper, _ abci.RequestQuery, cdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) != 1 {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "data hash not specified")
	}
	return k.fileCache.GetFile(path[0])
}

func queryDataSources(ctx sdk.Context, path []string, k Keeper, req abci.RequestQuery, cdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) != 0 {
		return queryDataSourceByID(ctx, path, k, req, cdc)
	}
	var params commontypes.QueryPaginationParams
	if err := cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	ctx, _ = ctx.CacheContext()
	dataSources, pageRes, err := k.GetPaginatedDataSources(ctx, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}
	return commontypes.QueryOK(cdc, oracletypes.QueryDataSourcesResponse{DataSources: dataSources, Pagination: pageRes})
}

func queryDataSourceByID(ctx sdk.Context, path []string, k Keeper, _ abci.RequestQuery, cdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) != 1 {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "data source not specified")
	}
	id, err := strconv.ParseInt(path[0], 10, 64)
	if err != nil {
		return commontypes.QueryBadRequest(cdc, err.Error())
	}
	dataSource, err := k.GetDataSource(ctx, oracletypes.DataSourceID(id))
	if err != nil {
		return commontypes.QueryNotFound(cdc, err.Error())
	}
	return commontypes.QueryOK(cdc, dataSource)
}

func queryOracleScripts(ctx sdk.Context, path []string, k Keeper, req abci.RequestQuery, cdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) != 0 {
		return queryOracleScriptByID(ctx, path, k, req, cdc)
	}
	var params commontypes.QueryPaginationParams
	if err := cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	ctx, _ = ctx.CacheContext()
	oracleScripts, pageRes, err := k.GetPaginatedOracleScripts(ctx, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}
	return commontypes.QueryOK(cdc, oracletypes.QueryOracleScriptsResponse{OracleScripts: oracleScripts, Pagination: pageRes})
}

func queryOracleScriptByID(ctx sdk.Context, path []string, k Keeper, _ abci.RequestQuery, cdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) != 1 {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "oracle script not specified")
	}
	id, err := strconv.ParseInt(path[0], 10, 64)
	if err != nil {
		return commontypes.QueryBadRequest(cdc, err.Error())
	}
	oracleScript, err := k.GetOracleScript(ctx, oracletypes.OracleScriptID(id))
	if err != nil {
		return commontypes.QueryNotFound(cdc, err.Error())
	}
	return commontypes.QueryOK(cdc, oracleScript)
}

func queryRequests(ctx sdk.Context, path []string, k Keeper, req abci.RequestQuery, cdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) != 0 {
		return queryRequestByID(ctx, path, k, req, cdc)
	}
	var params commontypes.QueryPaginationParams
	if err := cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	ctx, _ = ctx.CacheContext()
	requests, pageRes, err := k.GetPaginatedRequests(ctx, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}
	return commontypes.QueryOK(cdc, oracletypes.QueryRequestsResponse{Requests: requests, Pagination: pageRes})
}

func queryRequestByID(ctx sdk.Context, path []string, k Keeper, _ abci.RequestQuery, cdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) < 1 {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "request not specified")
	}
	id, err := strconv.ParseInt(path[0], 10, 64)
	if err != nil {
		return commontypes.QueryBadRequest(cdc, err.Error())
	}
	result, err := k.GetResult(ctx, oracletypes.RequestID(id))
	if err != nil {
		return nil, err
	}
	request := &oracletypes.RequestResult{
		RequestPacketData: &oracletypes.OracleRequestPacketData{
			ClientID:       result.ClientID,
			OracleScriptID: result.OracleScriptID,
			Calldata:       result.Calldata,
			AskCount:       result.AskCount,
			MinCount:       result.MinCount,
		},
		ResponsePacketData: &oracletypes.OracleResponsePacketData{
			RequestID:     result.RequestID,
			AnsCount:      result.AnsCount,
			RequestTime:   result.RequestTime,
			ResolveTime:   result.ResolveTime,
			ResolveStatus: result.ResolveStatus,
			Result:        result.Result,
		},
	}
	return commontypes.QueryOK(cdc, oracletypes.QueryRequestResponse{Request: request})
}

func queryValidatorStatus(ctx sdk.Context, path []string, k Keeper, _ abci.RequestQuery, cdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) != 1 {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "validator address not specified")
	}
	validatorAddress, err := sdk.ValAddressFromBech32(path[0])
	if err != nil {
		return commontypes.QueryBadRequest(cdc, err.Error())
	}
	return commontypes.QueryOK(cdc, k.GetValidatorStatus(ctx, validatorAddress))
}

func queryReporters(ctx sdk.Context, path []string, k Keeper, _ abci.RequestQuery, cdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) != 1 {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "validator address not specified")
	}
	validatorAddress, err := sdk.ValAddressFromBech32(path[0])
	if err != nil {
		return commontypes.QueryBadRequest(cdc, err.Error())
	}
	return commontypes.QueryOK(cdc, k.GetReporters(ctx, validatorAddress))
}

func queryActiveValidators(ctx sdk.Context, k Keeper, _ abci.RequestQuery, cdc *codec.LegacyAmino) ([]byte, error) {
	var vals []oracletypes.QueryActiveValidatorResult
	k.stakingKeeper.IterateBondedValidatorsByPower(ctx,
		func(idx int64, val stakingtypes.ValidatorI) (stop bool) {
			if k.GetValidatorStatus(ctx, val.GetOperator()).IsActive {
				vals = append(vals, oracletypes.QueryActiveValidatorResult{
					Address: val.GetOperator(),
					Power:   val.GetTokens().Uint64(),
				})
			}
			return false
		})
	return commontypes.QueryOK(cdc, vals)
}

func queryPendingRequests(ctx sdk.Context, path []string, k Keeper, _ abci.RequestQuery, cdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) > 1 {
		return commontypes.QueryBadRequest(cdc, "too many arguments")
	}

	var valAddress *sdk.ValAddress
	if len(path) == 1 {
		valAddress = new(sdk.ValAddress)
		address, err := sdk.ValAddressFromBech32(path[0])
		if err != nil {
			return commontypes.QueryBadRequest(cdc, err.Error())
		}

		*valAddress = address
	}

	lastExpired := k.GetRequestLastExpired(ctx)
	requestCount := k.GetRequestCount(ctx)

	var pendingIDs []oracletypes.RequestID
	for id := lastExpired + 1; int64(id) <= requestCount; id++ {

		req := k.MustGetRequest(ctx, id)

		// If all validators reported on this request, then skip it.
		reports := k.GetRequestReports(ctx, id)
		if len(reports) == len(req.RequestedValidators) {
			continue
		}

		// Skip if validator hasn't been assigned.
		if valAddress != nil {

			// If the validator isn't in requested validators set, then skip it.
			isValidator := false
			for _, v := range req.RequestedValidators {
				valAddr, err := sdk.ValAddressFromBech32(v)
				if err != nil {
					return commontypes.QueryBadRequest(cdc, err.Error())
				}
				if valAddress.Equals(valAddr) {
					isValidator = true
					break
				}
			}

			if !isValidator {
				continue
			}

			// If the validator has reported, then skip it.
			reported := false
			for _, r := range reports {
				valAddr, err := sdk.ValAddressFromBech32(r.Validator)
				if err != nil {
					return commontypes.QueryBadRequest(cdc, err.Error())
				}
				if valAddress.Equals(valAddr) {
					reported = true
					break
				}
			}

			if reported {
				continue
			}
		}

		pendingIDs = append(pendingIDs, id)
	}

	return commontypes.QueryOK(cdc, pendingIDs)
}

func queryRequestReports(ctx sdk.Context, path []string, k Keeper, req abci.RequestQuery, cdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) != 1 {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "not enough arguments")
	}
	requestId, err := strconv.ParseInt(path[0], 10, 64)
	if err != nil {
		return commontypes.QueryBadRequest(cdc, err.Error())
	}
	var params commontypes.QueryPaginationParams
	if err := cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	ctx, _ = ctx.CacheContext()
	reports, pageRes, err := k.GetPaginatedRequestReports(ctx, oracletypes.RequestID(requestId), params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}
	return commontypes.QueryOK(cdc, oracletypes.QueryRequestReportsResponse{Reports: reports, Pagination: pageRes})
}

func queryDataProvidersPool(ctx sdk.Context, k Keeper, _ abci.RequestQuery, cdc *codec.LegacyAmino) ([]byte, error) {
	return commontypes.QueryOK(cdc, k.GetOraclePool(ctx).DataProvidersPool)
}

func queryDataProvidersRewardPerByte(ctx sdk.Context, k Keeper, _ abci.RequestQuery, cdc *codec.LegacyAmino) ([]byte, error) {
	accumulatedRewards := k.GetAccumulatedDataProvidersRewards(ctx)
	return commontypes.QueryOK(cdc, accumulatedRewards.CurrentRewardPerByte)
}
