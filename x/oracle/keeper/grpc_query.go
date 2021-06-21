package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// Counts queries the number of data sources, oracle scripts, and requests.
func (k Querier) Counts(c context.Context, req *types.QueryCountsRequest) (*types.QueryCountsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryCountsResponse{
			DataSourceCount:   k.GetDataSourceCount(ctx),
			OracleScriptCount: k.GetOracleScriptCount(ctx),
			RequestCount:      k.GetRequestCount(ctx)},
		nil
}

// Data queries the data source or oracle script script for given file hash.
func (k Querier) Data(c context.Context, req *types.QueryDataRequest) (*types.QueryDataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	data, err := k.fileCache.GetFile(req.DataHash)
	if err != nil {
		return nil, err
	}
	return &types.QueryDataResponse{Data: data}, nil
}

// DataSource queries data source info for given data source id.
func (k Querier) DataSource(c context.Context, req *types.QueryDataSourceRequest) (*types.QueryDataSourceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	ds, err := k.GetDataSource(ctx, types.DataSourceID(req.DataSourceId))
	if err != nil {
		return nil, err
	}
	return &types.QueryDataSourceResponse{DataSource: &ds}, nil
}

// OracleScript queries oracle script info for given oracle script id.
func (k Querier) OracleScript(c context.Context, req *types.QueryOracleScriptRequest) (*types.QueryOracleScriptResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	os, err := k.GetOracleScript(ctx, types.OracleScriptID(req.OracleScriptId))
	if err != nil {
		return nil, err
	}
	return &types.QueryOracleScriptResponse{OracleScript: &os}, nil
}

// Request queries request info for given request id.
func (k Querier) Request(c context.Context, req *types.QueryRequestRequest) (*types.QueryRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	rid := types.RequestID(req.RequestId)

	request, err := k.GetRequest(ctx, rid)
	if err != nil {
		lastExpired := k.GetRequestLastExpired(ctx)
		if rid > lastExpired {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to get request from chain: request id (%d) > latest expired request id (%d)", rid, lastExpired))
		}
		result := k.MustGetResult(ctx, rid)
		return &types.QueryRequestResponse{Request: nil, Reports: nil, Result: &result}, nil
	}

	reports := k.GetReports(ctx, rid)
	if !k.HasResult(ctx, rid) {
		return &types.QueryRequestResponse{Request: &request, Reports: reports, Result: nil}, nil
	}

	result := k.MustGetResult(ctx, rid)
	return &types.QueryRequestResponse{Request: &request, Reports: reports, Result: &result}, nil
}

func (k Querier) PendingRequests(c context.Context, req *types.QueryPendingRequestsRequest) (*types.QueryPendingRequestsResponse, error) {
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
		reports := k.GetReports(ctx, id)
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

	return &types.QueryPendingRequestsResponse{RequestIDs: pendingIDs}, nil
}

// Validator queries oracle info of validator for given validator
// address.
func (k Querier) Validator(c context.Context, req *types.QueryValidatorRequest) (*types.QueryValidatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	val, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	status := k.GetValidatorStatus(ctx, val)
	if err != nil {
		return nil, err
	}
	return &types.QueryValidatorResponse{Status: &status}, nil
}

// Reporters queries all reporters of a given validator address.
func (k Querier) Reporters(c context.Context, req *types.QueryReportersRequest) (*types.QueryReportersResponse, error) {
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
	return &types.QueryReportersResponse{Reporter: reporters}, nil
}

// ActiveValidators queries all active oracle validators.
func (k Querier) ActiveValidators(c context.Context, req *types.QueryActiveValidatorsRequest) (*types.QueryActiveValidatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	result := types.QueryActiveValidatorsResponse{}
	k.stakingKeeper.IterateBondedValidatorsByPower(ctx,
		func(idx int64, val stakingtypes.ValidatorI) (stop bool) {
			if k.GetValidatorStatus(ctx, val.GetOperator()).IsActive {
				result.Validators = append(result.Validators, &types.ActiveValidator{
					Address: val.GetOperator().String(),
					Power:   val.GetTokens().Uint64(),
				})
			}
			return false
		})
	return &result, nil
}

// Params queries the oracle parameters.
func (k Querier) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}

// RequestSearch queries the latest request that match the given input.
func (k Querier) RequestSearch(c context.Context, req *types.QueryRequestSearchRequest) (*types.QueryRequestSearchResponse, error) {
	return nil, status.Error(codes.Unimplemented, "This feature can be taken from extra/rest branch")
}

// RequestPrice queries the latest price on standard price reference oracle
// script.
func (k Querier) RequestPrice(c context.Context, req *types.QueryRequestPriceRequest) (*types.QueryRequestPriceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "This feature can be taken from extra/rest branch")
}

// RequestVerification verifies oracle request for validation before executing data sources
func (k Querier) RequestVerification(c context.Context, req *types.QueryRequestVerificationRequest) (*types.QueryRequestVerificationResponse, error) {
	// Request should not be empty
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// Provided chain ID should match current chain ID
	if ctx.ChainID() != req.ChainId {
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("provided chain ID does not match the validator's chain ID; expected %s, got %s", ctx.ChainID(), req.ChainId))
	}

	// Provided validator's address should be valid
	validator, err := sdk.ValAddressFromBech32(req.Validator)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("unable to parse validator address: %s", err.Error()))
	}

	// Provided signature should be valid, which means this query request should be signed by the provided reporter
	reporterPubKey, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeAccPub, req.Reporter)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("unable to get reporter's public key: %s", err.Error()))
	}
	requestVerificationContent := types.NewRequestVerification(req.ChainId, validator, types.RequestID(req.RequestId), types.ExternalID(req.ExternalId))
	signByte := requestVerificationContent.GetSignBytes()
	if !reporterPubKey.VerifySignature(signByte, req.Signature) {
		return nil, status.Error(codes.Unauthenticated, "invalid reporter's signature")
	}

	// Provided reporter should be authorized by the provided validator
	reporters := k.GetReporters(ctx, validator)
	reporter := sdk.AccAddress(reporterPubKey.Address().Bytes())
	isReporterAuthorizedByValidator := false
	for _, existingReporter := range reporters {
		if reporter.Equals(existingReporter) {
			isReporterAuthorizedByValidator = true
			break
		}
	}
	if !isReporterAuthorizedByValidator {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("%s is not an authorized reporter of %s", reporter, req.Validator))
	}

	// Provided request should exist on chain
	request, err := k.GetRequest(ctx, types.RequestID(req.RequestId))
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to get request from chain: %s", err.Error()))
	}

	// Provided validator should be assigned to response to the request
	isValidatorAssigned := false
	for _, requestedValidator := range request.RequestedValidators {
		v, _ := sdk.ValAddressFromBech32(requestedValidator)
		if validator.Equals(v) {
			isValidatorAssigned = true
			break
		}
	}
	if !isValidatorAssigned {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("%s is not assigned for request ID %d", validator, req.RequestId))
	}

	// Provided external ID should be required by the request determined by oracle script
	var dataSourceID *types.DataSourceID
	for _, rawRequest := range request.RawRequests {
		if rawRequest.ExternalID == types.ExternalID(req.ExternalId) {
			dataSourceID = &rawRequest.DataSourceID
			break
		}
	}
	if dataSourceID == nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("no data source required by the request %d found which relates to the external data source with ID %d.", req.RequestId, req.ExternalId))
	}

	// Provided validator should not have reported data for the request
	reports := k.GetReports(ctx, types.RequestID(req.RequestId))
	isValidatorReported := false
	for _, report := range reports {
		reportVal, _ := sdk.ValAddressFromBech32(report.Validator)
		if reportVal.Equals(validator) {
			isValidatorReported = true
			break
		}
	}
	if isValidatorReported {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("validator %s already submitted data report for this request", validator))
	}

	// The request should not be expired
	if request.RequestHeight+int64(k.ExpirationBlockCount(ctx)) < ctx.BlockHeader().Height {
		return nil, status.Error(codes.DeadlineExceeded, fmt.Sprintf("Request with ID %d is already expired", req.RequestId))
	}

	return &types.QueryRequestVerificationResponse{
		ChainId:      req.ChainId,
		Validator:    req.Validator,
		RequestId:    req.RequestId,
		ExternalId:   req.ExternalId,
		DataSourceId: int64(*dataSourceID),
	}, nil
}

// RequestPool queries the request pool information
func (k Querier) RequestPool(c context.Context, req *types.QueryRequestPoolRequest) (*types.QueryRequestPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	requestPool := types.GetEscrowAddress(req.RequestKey, req.PortId, req.ChannelId)
	b := k.bankKeeper.GetAllBalances(ctx, requestPool)
	return &types.QueryRequestPoolResponse{RequestPoolAddress: requestPool.String(), Balance: b}, nil
}
