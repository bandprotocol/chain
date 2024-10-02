package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

func (suite *KeeperTestSuite) TestResultBasicFunctions() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// We start by setting result of request#1.
	result := types.NewResult(
		"alice", 1, basicCalldata, 1, 1, 1, 1, 1589535020, 1589535022, 1, basicResult,
	)
	k.SetResult(ctx, 1, result)
	// GetResult and MustGetResult should return what we set.
	result, err := k.GetResult(ctx, 1)
	require.NoError(err)
	require.Equal(result, result)
	result = k.MustGetResult(ctx, 1)
	require.Equal(result, result)
	// GetResult of another request should return error.
	_, err = k.GetResult(ctx, 2)
	require.ErrorIs(err, types.ErrResultNotFound)
	require.Panics(func() { k.MustGetResult(ctx, 2) })
	// HasResult should also perform correctly.
	require.True(k.HasResult(ctx, 1))
	require.False(k.HasResult(ctx, 2))
}

func (suite *KeeperTestSuite) TestSaveResultOK() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	ctx = ctx.WithBlockTime(bandtesting.ParseTime(200))
	k.SetRequest(ctx, 42, defaultRequest()) // See report_test.go
	k.SetReport(ctx, 42, types.NewReport(validators[0].Address, true, nil))
	k.SaveResult(ctx, 42, types.RESOLVE_STATUS_SUCCESS, basicResult)
	expect := types.NewResult(
		basicClientID, 1, basicCalldata, 2, 2, 42, 1, bandtesting.ParseTime(0).Unix(),
		bandtesting.ParseTime(200).Unix(), types.RESOLVE_STATUS_SUCCESS, basicResult,
	)
	result, err := k.GetResult(ctx, 42)
	require.NoError(err)
	require.Equal(expect, result)
}

func (suite *KeeperTestSuite) TestResolveSuccess() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	k.SetRequest(ctx, 42, defaultRequest()) // See report_test.go
	k.SetReport(ctx, 42, types.NewReport(validators[0].Address, true, nil))
	k.ResolveSuccess(ctx, 42, basicResult, 1234)
	require.Equal(types.RESOLVE_STATUS_SUCCESS, k.MustGetResult(ctx, 42).ResolveStatus)
	require.Equal(basicResult, k.MustGetResult(ctx, 42).Result)
	require.Equal(sdk.Events{sdk.NewEvent(
		types.EventTypeResolve,
		sdk.NewAttribute(types.AttributeKeyID, "42"),
		sdk.NewAttribute(types.AttributeKeyResolveStatus, "1"),
		sdk.NewAttribute(types.AttributeKeyResult, "42415349435f524553554c54"), // BASIC_RESULT
		sdk.NewAttribute(types.AttributeKeyGasUsed, "1234"),
	)}, ctx.EventManager().Events())
}

func (suite *KeeperTestSuite) TestResolveFailure() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	k.SetRequest(ctx, 42, defaultRequest()) // See report_test.go
	k.SetReport(ctx, 42, types.NewReport(validators[0].Address, true, nil))
	k.ResolveFailure(ctx, 42, "REASON")
	require.Equal(types.RESOLVE_STATUS_FAILURE, k.MustGetResult(ctx, 42).ResolveStatus)
	require.Empty(k.MustGetResult(ctx, 42).Result)
	require.Equal(sdk.Events{sdk.NewEvent(
		types.EventTypeResolve,
		sdk.NewAttribute(types.AttributeKeyID, "42"),
		sdk.NewAttribute(types.AttributeKeyResolveStatus, "2"),
		sdk.NewAttribute(types.AttributeKeyReason, "REASON"),
	)}, ctx.EventManager().Events())
}

func (suite *KeeperTestSuite) TestResolveExpired() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	k.SetRequest(ctx, 42, defaultRequest()) // See report_test.go
	k.SetReport(ctx, 42, types.NewReport(validators[0].Address, true, nil))
	k.ResolveExpired(ctx, 42)
	require.Equal(types.RESOLVE_STATUS_EXPIRED, k.MustGetResult(ctx, 42).ResolveStatus)
	require.Empty(k.MustGetResult(ctx, 42).Result)
	require.Equal(sdk.Events{sdk.NewEvent(
		types.EventTypeResolve,
		sdk.NewAttribute(types.AttributeKeyID, "42"),
		sdk.NewAttribute(types.AttributeKeyResolveStatus, "3"),
	)}, ctx.EventManager().Events())
}
