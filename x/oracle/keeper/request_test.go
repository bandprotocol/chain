package keeper_test

import (
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/oracle/keeper"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

func testRequest(
	require *require.Assertions,
	k keeper.Keeper,
	ctx sdk.Context,
	rid types.RequestID,
	resolveStatus types.ResolveStatus,
	reportCount uint64,
	hasRequest bool,
) {
	if resolveStatus == types.RESOLVE_STATUS_OPEN {
		require.False(k.HasResult(ctx, rid))
	} else {
		r, err := k.GetResult(ctx, rid)
		require.NoError(err)
		require.NotNil(r)
		require.Equal(resolveStatus, r.ResolveStatus)
	}

	require.Equal(reportCount, k.GetReportCount(ctx, rid))
	require.Equal(hasRequest, k.HasRequest(ctx, rid))
}

func (suite *KeeperTestSuite) TestHasRequest() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// We should not have a request ID 42 without setting it.
	require.False(k.HasRequest(ctx, 42))
	// After we set it, we should be able to find it.
	k.SetRequest(ctx, 42, types.NewRequest(1, basicCalldata, nil, 1, 1, bandtesting.ParseTime(0), "", nil, nil, 0))
	require.True(k.HasRequest(ctx, 42))
}

func (suite *KeeperTestSuite) TestDeleteRequest() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// After we set it, we should be able to find it.
	k.SetRequest(ctx, 42, types.NewRequest(1, basicCalldata, nil, 1, 1, bandtesting.ParseTime(0), "", nil, nil, 0))
	require.True(k.HasRequest(ctx, 42))
	// After we delete it, we should not find it anymore.
	k.DeleteRequest(ctx, 42)
	require.False(k.HasRequest(ctx, 42))
	_, err := k.GetRequest(ctx, 42)
	require.ErrorIs(err, types.ErrRequestNotFound)
	require.Panics(func() { _ = k.MustGetRequest(ctx, 42) })
}

func (suite *KeeperTestSuite) TestSetterGetterRequest() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// Getting a non-existent request should return error.
	_, err := k.GetRequest(ctx, 42)
	require.ErrorIs(err, types.ErrRequestNotFound)
	require.Panics(func() { _ = k.MustGetRequest(ctx, 42) })
	// Creates some basic requests.
	req1 := types.NewRequest(1, basicCalldata, nil, 1, 1, bandtesting.ParseTime(0), "", nil, nil, 0)
	req2 := types.NewRequest(2, basicCalldata, nil, 1, 1, bandtesting.ParseTime(0), "", nil, nil, 0)
	// Sets id 42 with request 1 and id 42 with request 2.
	k.SetRequest(ctx, 42, req1)
	k.SetRequest(ctx, 43, req2)
	// Checks that Get and MustGet perform correctly.
	req1Res, err := k.GetRequest(ctx, 42)
	require.Nil(err)
	require.Equal(req1, req1Res)
	require.Equal(req1, k.MustGetRequest(ctx, 42))
	req2Res, err := k.GetRequest(ctx, 43)
	require.Nil(err)
	require.Equal(req2, req2Res)
	require.Equal(req2, k.MustGetRequest(ctx, 43))
	// Replaces id 42 with another request.
	k.SetRequest(ctx, 42, req2)
	require.NotEqual(req1, k.MustGetRequest(ctx, 42))
	require.Equal(req2, k.MustGetRequest(ctx, 42))
}

func (suite *KeeperTestSuite) TestSetterGettterPendingResolveList() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// Initially, we should get an empty list of pending resolves.
	require.Equal(k.GetPendingResolveList(ctx), []types.RequestID{})
	// After we set something, we should get that thing back.
	k.SetPendingResolveList(ctx, []types.RequestID{5, 6, 7, 8})
	require.Equal(k.GetPendingResolveList(ctx), []types.RequestID{5, 6, 7, 8})
	// Let's also try setting it back to empty list.
	k.SetPendingResolveList(ctx, []types.RequestID{})
	require.Equal(k.GetPendingResolveList(ctx), []types.RequestID{})
	// Nil should also works.
	k.SetPendingResolveList(ctx, nil)
	require.Equal(k.GetPendingResolveList(ctx), []types.RequestID{})
}

func (suite *KeeperTestSuite) TestAddDataSourceBasic() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// We start by setting an oracle request available at ID 42.
	k.SetOracleScript(ctx, 42, types.NewOracleScript(
		owner, basicName, basicDesc, basicFilename, basicSchema, basicSourceCodeURL,
	))
	// Adding the first request should return ID 1.
	id := k.AddRequest(
		ctx,
		types.NewRequest(42, basicCalldata, []sdk.ValAddress{}, 1, 1, bandtesting.ParseTime(0), "", nil, nil, 0),
	)
	require.Equal(id, types.RequestID(1))
	// Adding another request should return ID 2.
	id = k.AddRequest(
		ctx,
		types.NewRequest(42, basicCalldata, []sdk.ValAddress{}, 1, 1, bandtesting.ParseTime(0), "", nil, nil, 0),
	)
	require.Equal(id, types.RequestID(2))
}

func (suite *KeeperTestSuite) TestAddPendingResolveList() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// Initially, we should get an empty list of pending resolves.
	require.Equal(k.GetPendingResolveList(ctx), []types.RequestID{})
	// Everytime we append a new request ID, it should show up.
	k.AddPendingRequest(ctx, 42)
	require.Equal(k.GetPendingResolveList(ctx), []types.RequestID{42})
	k.AddPendingRequest(ctx, 43)
	require.Equal(k.GetPendingResolveList(ctx), []types.RequestID{42, 43})
}

func (suite *KeeperTestSuite) TestProcessExpiredRequests() {
	suite.activeAllValidators()
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	params := k.GetParams(ctx)
	params.ExpirationBlockCount = 3
	err := k.SetParams(ctx, params)
	require.NoError(err)

	// Set some initial requests. All requests are asked to validators 0 & 1.
	// All request time is set to 1 second after the validators are activated.
	req1 := defaultRequest()
	req1.RequestHeight = 5
	req1.RequestTime = ctx.BlockHeader().Time.Unix() + 1
	req2 := defaultRequest()
	req2.RequestHeight = 6
	req2.RequestTime = ctx.BlockHeader().Time.Unix() + 1
	req3 := defaultRequest()
	req3.RequestHeight = 6
	req3.RequestTime = ctx.BlockHeader().Time.Unix() + 1
	req4 := defaultRequest()
	req4.RequestHeight = 10
	req4.RequestTime = ctx.BlockHeader().Time.Unix() + 1
	k.AddRequest(ctx, req1)
	k.AddRequest(ctx, req2)
	k.AddRequest(ctx, req3)
	k.AddRequest(ctx, req4)

	// Initially validator 0 & 1 are active.
	require.True(k.GetValidatorStatus(ctx, validators[0].Address).IsActive)
	require.True(k.GetValidatorStatus(ctx, validators[1].Address).IsActive)

	// Validator 1 reports all requests. Validator 2 misses request#3.
	rawReports := []types.RawReport{
		types.NewRawReport(1, 0, []byte("data1/1")),
		types.NewRawReport(2, 1, []byte("data2/1")),
		types.NewRawReport(3, 0, []byte("data3/1")),
	}
	err = k.AddReport(ctx, 1, validators[0].Address, false, rawReports)
	require.NoError(err)
	err = k.AddReport(ctx, 2, validators[0].Address, true, rawReports)
	require.NoError(err)
	err = k.AddReport(ctx, 3, validators[0].Address, false, rawReports)
	require.NoError(err)
	err = k.AddReport(ctx, 4, validators[0].Address, true, rawReports)
	require.NoError(err)
	err = k.AddReport(ctx, 1, validators[1].Address, true, rawReports)
	require.NoError(err)
	err = k.AddReport(ctx, 2, validators[1].Address, true, rawReports)
	require.NoError(err)
	err = k.AddReport(ctx, 4, validators[1].Address, true, rawReports)
	require.NoError(err)

	// Request 1, 2 and 4 gets resolved. Request 3 does not.
	k.ResolveSuccess(ctx, 1, basicResult, 1234)
	k.ResolveFailure(ctx, 2, "ARBITRARY_REASON")
	k.ResolveSuccess(ctx, 4, basicResult, 1234)
	// Initially, last expired request ID should be 0.
	require.Equal(types.RequestID(0), k.GetRequestLastExpired(ctx))

	// At block 7, nothing should happen.
	ctx = ctx.WithBlockHeight(7).WithBlockTime(bandtesting.ParseTime(7000)).WithEventManager(sdk.NewEventManager())
	k.ProcessExpiredRequests(ctx)
	require.Equal(sdk.Events{}, ctx.EventManager().Events())
	require.Equal(types.RequestID(0), k.GetRequestLastExpired(ctx))
	require.True(k.GetValidatorStatus(ctx, validators[0].Address).IsActive)
	require.True(k.GetValidatorStatus(ctx, validators[1].Address).IsActive)
	testRequest(require, k, ctx, types.RequestID(1), types.RESOLVE_STATUS_SUCCESS, 2, true)
	testRequest(require, k, ctx, types.RequestID(2), types.RESOLVE_STATUS_FAILURE, 2, true)
	testRequest(require, k, ctx, types.RequestID(3), types.RESOLVE_STATUS_OPEN, 1, true)
	testRequest(require, k, ctx, types.RequestID(4), types.RESOLVE_STATUS_SUCCESS, 2, true)

	// At block 8, now last request ID should move to 1. No events should be emitted.
	ctx = ctx.WithBlockHeight(8).WithBlockTime(bandtesting.ParseTime(8000)).WithEventManager(sdk.NewEventManager())
	k.ProcessExpiredRequests(ctx)
	require.Equal(sdk.Events{}, ctx.EventManager().Events())
	require.Equal(types.RequestID(1), k.GetRequestLastExpired(ctx))
	require.True(k.GetValidatorStatus(ctx, validators[0].Address).IsActive)
	require.True(k.GetValidatorStatus(ctx, validators[1].Address).IsActive)
	testRequest(require, k, ctx, types.RequestID(1), types.RESOLVE_STATUS_SUCCESS, 0, false)
	testRequest(require, k, ctx, types.RequestID(2), types.RESOLVE_STATUS_FAILURE, 2, true)
	testRequest(require, k, ctx, types.RequestID(3), types.RESOLVE_STATUS_OPEN, 1, true)
	testRequest(require, k, ctx, types.RequestID(4), types.RESOLVE_STATUS_SUCCESS, 2, true)

	// At block 9, request#3 is expired and validator 2 becomes inactive.
	ctx = ctx.WithBlockHeight(9).WithBlockTime(bandtesting.ParseTime(9000)).WithEventManager(sdk.NewEventManager())
	k.ProcessExpiredRequests(ctx)
	require.Equal(sdk.Events{sdk.NewEvent(
		types.EventTypeResolve,
		sdk.NewAttribute(types.AttributeKeyID, "3"),
		sdk.NewAttribute(types.AttributeKeyResolveStatus, "3"),
	), sdk.NewEvent(
		types.EventTypeDeactivate,
		sdk.NewAttribute(types.AttributeKeyValidator, validators[1].Address.String()),
	)}, ctx.EventManager().Events())
	require.Equal(types.RequestID(3), k.GetRequestLastExpired(ctx))
	require.True(k.GetValidatorStatus(ctx, validators[0].Address).IsActive)
	require.False(k.GetValidatorStatus(ctx, validators[1].Address).IsActive)
	require.Equal(types.NewResult(
		basicClientID, req3.OracleScriptID, req3.Calldata, uint64(len(req3.RequestedValidators)), req3.MinCount,
		3, 1, req3.RequestTime, bandtesting.ParseTime(9000).Unix(),
		types.RESOLVE_STATUS_EXPIRED, nil,
	), k.MustGetResult(ctx, 3))
	testRequest(require, k, ctx, types.RequestID(1), types.RESOLVE_STATUS_SUCCESS, 0, false)
	testRequest(require, k, ctx, types.RequestID(2), types.RESOLVE_STATUS_FAILURE, 0, false)
	testRequest(require, k, ctx, types.RequestID(3), types.RESOLVE_STATUS_EXPIRED, 0, false)
	testRequest(require, k, ctx, types.RequestID(4), types.RESOLVE_STATUS_SUCCESS, 2, true)

	// At block 10, nothing should happen
	ctx = ctx.WithBlockHeight(10).WithBlockTime(bandtesting.ParseTime(10000)).WithEventManager(sdk.NewEventManager())
	k.ProcessExpiredRequests(ctx)
	require.Equal(sdk.Events{}, ctx.EventManager().Events())
	require.Equal(types.RequestID(3), k.GetRequestLastExpired(ctx))
	testRequest(require, k, ctx, types.RequestID(1), types.RESOLVE_STATUS_SUCCESS, 0, false)
	testRequest(require, k, ctx, types.RequestID(2), types.RESOLVE_STATUS_FAILURE, 0, false)
	testRequest(require, k, ctx, types.RequestID(3), types.RESOLVE_STATUS_EXPIRED, 0, false)
	testRequest(require, k, ctx, types.RequestID(4), types.RESOLVE_STATUS_SUCCESS, 2, true)

	// At block 13, last expired request becomes 4.
	ctx = ctx.WithBlockHeight(13).WithBlockTime(bandtesting.ParseTime(13000)).WithEventManager(sdk.NewEventManager())
	k.ProcessExpiredRequests(ctx)
	require.Equal(sdk.Events{}, ctx.EventManager().Events())
	require.Equal(types.RequestID(4), k.GetRequestLastExpired(ctx))
	testRequest(require, k, ctx, types.RequestID(1), types.RESOLVE_STATUS_SUCCESS, 0, false)
	testRequest(require, k, ctx, types.RequestID(2), types.RESOLVE_STATUS_FAILURE, 0, false)
	testRequest(require, k, ctx, types.RequestID(3), types.RESOLVE_STATUS_EXPIRED, 0, false)
	testRequest(require, k, ctx, types.RequestID(4), types.RESOLVE_STATUS_SUCCESS, 0, false)
}
