package oraclekeeper

import (
	"context"
	"fmt"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	Keeper
}

var _ oracletypes.QueryServer = Querier{}

// Counts queries the number of data sources, oracle scripts, and requests.
func (k Querier) Counts(c context.Context, req *oracletypes.QueryCountsRequest) (*oracletypes.QueryCountsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &oracletypes.QueryCountsResponse{
			DataSourceCount:   k.GetDataSourceCount(ctx),
			OracleScriptCount: k.GetOracleScriptCount(ctx),
			RequestCount:      k.GetRequestCount(ctx)},
		nil
}

// Data queries the data source or oracle script script for given file hash.
func (k Querier) Data(c context.Context, req *oracletypes.QueryDataRequest) (*oracletypes.QueryDataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	data, err := k.fileCache.GetFile(req.DataHash)
	if err != nil {
		return nil, err
	}
	return &oracletypes.QueryDataResponse{Data: data}, nil
}

// DataSource queries data source info for given data source id.
func (k Querier) DataSource(c context.Context, req *oracletypes.QueryDataSourceRequest) (*oracletypes.QueryDataSourceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	ds, err := k.GetDataSource(ctx, oracletypes.DataSourceID(req.DataSourceId))
	if err != nil {
		return nil, err
	}
	return &oracletypes.QueryDataSourceResponse{DataSource: &ds}, nil
}

// DataSources queries data sources
func (k Querier) DataSources(c context.Context, req *oracletypes.QueryDataSourcesRequest) (*oracletypes.QueryDataSourcesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	dataSources, pageRes, err := k.GetPaginatedDataSources(ctx, req.Pagination.Limit, req.Pagination.Offset)
	if err != nil {
		return nil, err
	}
	return &oracletypes.QueryDataSourcesResponse{DataSources: dataSources, Pagination: pageRes}, nil
}

// OracleScript queries oracle script info for given oracle script id.
func (k Querier) OracleScript(c context.Context, req *oracletypes.QueryOracleScriptRequest) (*oracletypes.QueryOracleScriptResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	os, err := k.GetOracleScript(ctx, oracletypes.OracleScriptID(req.OracleScriptId))
	if err != nil {
		return nil, err
	}
	return &oracletypes.QueryOracleScriptResponse{OracleScript: &os}, nil
}

// OracleScripts queries all oracle scripts with pagination.
func (k Querier) OracleScripts(c context.Context, req *oracletypes.QueryOracleScriptsRequest) (*oracletypes.QueryOracleScriptsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	oracleScripts, pageRes, err := k.GetPaginatedOracleScripts(ctx, req.Pagination.Limit, req.Pagination.Offset)
	if err != nil {
		return nil, err
	}
	return &oracletypes.QueryOracleScriptsResponse{OracleScripts: oracleScripts, Pagination: pageRes}, nil
}

// Request queries request info for given request id.
func (k Querier) Request(c context.Context, req *oracletypes.QueryRequestRequest) (*oracletypes.QueryRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	result, err := k.GetResult(ctx, oracletypes.RequestID(req.RequestId))
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
	return &oracletypes.QueryRequestResponse{Request: request}, nil
}

// Requests queries all requests with pagination.
func (k Querier) Requests(c context.Context, req *oracletypes.QueryRequestsRequest) (*oracletypes.QueryRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	requests, pageRes, err := k.GetPaginatedRequests(ctx, req.Pagination.Limit, req.Pagination.Offset)
	if err != nil {
		return nil, err
	}
	return &oracletypes.QueryRequestsResponse{Requests: requests, Pagination: pageRes}, nil
}

// RequestReports queries all reports by the giver request id with pagination.
func (k Querier) RequestReports(c context.Context, req *oracletypes.QueryRequestReportsRequest) (*oracletypes.QueryRequestReportsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	reports, pageRes, err := k.GetPaginatedRequestReports(
		ctx,
		oracletypes.RequestID(req.RequestId),
		req.Pagination.Limit,
		req.Pagination.Offset,
	)
	if err != nil {
		return nil, err
	}
	return &oracletypes.QueryRequestReportsResponse{Reports: reports, Pagination: pageRes}, nil
}

// Validator queries oracle info of validator for given validator
// address.
func (k Querier) Validator(c context.Context, req *oracletypes.QueryValidatorRequest) (*oracletypes.QueryValidatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	val, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	validatorStatus := k.GetValidatorStatus(ctx, val)
	return &oracletypes.QueryValidatorResponse{Status: &validatorStatus}, nil
}

// Reporters queries all reporters of a given validator address.
func (k Querier) Reporters(c context.Context, req *oracletypes.QueryReportersRequest) (*oracletypes.QueryReportersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	val, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	reps := k.GetReporters(ctx, val)
	reporters := make([]string, len(reps))
	for idx, rep := range reps {
		reporters[idx] = rep.String()
	}
	return &oracletypes.QueryReportersResponse{Reporter: reporters}, nil
}

// ActiveValidators queries all active oracle validators.
func (k Querier) ActiveValidators(c context.Context, req *oracletypes.QueryActiveValidatorsRequest) (*oracletypes.QueryActiveValidatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
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
	return &oracletypes.QueryActiveValidatorsResponse{Count: int64(len(vals))}, nil
}

// Params queries the oracle parameters.
func (k Querier) Params(c context.Context, req *oracletypes.QueryParamsRequest) (*oracletypes.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)
	return &oracletypes.QueryParamsResponse{Params: params}, nil
}

// TODO: drop or change
// RequestSearch queries the latest request that matches the given input.
func (k Querier) RequestSearch(c context.Context, req *oracletypes.QueryRequestSearchRequest) (*oracletypes.QueryRequestSearchResponse, error) {

	// TODO: revisit, maybe find another way
	//var clientCtx client.Context
	//rawClientCtx := c.Value(client.ClientContextKey)
	//if rawClientCtx != nil {
	//	clientCtx = *rawClientCtx.(*client.Context)
	//} else {
	//	// SHOULD NEVER HIT
	//	panic("client ctx is empty")
	//}
	//clientCtx := client.Context{}
	//
	//resp, _, err := oracleclientcommon.QuerySearchLatestRequest(oracletypes.QuerierRoute, clientCtx, req)
	//if err != nil {
	//	return nil, err
	//}
	//
	//if resp == nil {
	//	return &oracletypes.QueryRequestSearchResponse{}, nil
	//}

	return nil, nil
}

// TODO:
// RequestPrice queries the latest price on standard price reference oracle script.
func (k Querier) RequestPrice(c context.Context, req *oracletypes.QueryRequestPriceRequest) (*oracletypes.QueryRequestPriceResponse, error) {
	return &oracletypes.QueryRequestPriceResponse{}, nil
}

func (k Querier) DataProvidersPool(c context.Context, req *oracletypes.QueryDataProvidersPoolRequest) (*oracletypes.QueryDataProvidersPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	return &oracletypes.QueryDataProvidersPoolResponse{
		Pool: k.GetOraclePool(ctx).DataProvidersPool,
	}, nil
}

// DataProviderReward returns current reward per byte for data providers
func (k Querier) DataProviderReward(
	c context.Context, _ *oracletypes.QueryDataProviderRewardRequest,
) (*oracletypes.QueryDataProviderRewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	accumulatedRewards := k.GetAccumulatedDataProvidersRewards(ctx)
	return &oracletypes.QueryDataProviderRewardResponse{RewardPerByte: accumulatedRewards.CurrentRewardPerByte}, nil
}


func (k Querier) PendingRequests(c context.Context, req *oracletypes.QueryPendingRequestsRequest) (*oracletypes.QueryPendingRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	valAddress, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("unable to parse given validator address: %v", err))
	}

	lastExpired := k.GetRequestLastExpired(ctx)
	requestCount := k.GetRequestCount(ctx)

	var pendingIDs []int64
	for id := lastExpired + 1; int64(id) <= requestCount; id++ {
		oracleReq := k.MustGetRequest(ctx, id)

		// If all validators reported on this request, then skip it.
		reports := k.GetRequestReports(ctx, id)
		if len(reports) == len(oracleReq.RequestedValidators) {
			continue
		}

		// Skip if validator hasn't been assigned or has been reported.
		// If the validator isn't in requested validators set, then skip it.
		isInValidatorSet := false
		for _, v := range oracleReq.RequestedValidators {
			val, err := sdk.ValAddressFromBech32(v)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("unable to parse validator address in requested validators %v: %v", v, err))
			}
			if valAddress.Equals(val) {
				isInValidatorSet = true
				break
			}
		}
		if !isInValidatorSet {
			continue
		}

		// If the validator has reported, then skip it.
		reported := false
		for _, r := range reports {
			val, err := sdk.ValAddressFromBech32(r.Validator)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("unable to parse validator address in requested validators %v: %v", r.Validator, err))
			}
			if valAddress.Equals(val) {
				reported = true
				break
			}
		}
		if reported {
			continue
		}

		pendingIDs = append(pendingIDs, int64(id))
	}

	return &oracletypes.QueryPendingRequestsResponse{RequestIDs: pendingIDs}, nil
}
