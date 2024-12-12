package keeper_test

import (
	"go.uber.org/mock/gomock"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
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
	k.ResolveSuccess(ctx, 42, defaultRequest().Requester, defaultRequest().FeeLimit, basicResult, 1234, 0)
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

func (suite *KeeperTestSuite) TestResolveSuccessButInsufficientMember() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()
	request := defaultRequest()
	request.TSSEncoder = types.ENCODER_FULL_ABI

	k.SetRequest(ctx, 42, request) // See report_test.go
	k.SetReport(ctx, 42, types.NewReport(validators[0].Address, true, nil))

	suite.bandtssKeeper.EXPECT().CreateDirectSigningRequest(
		gomock.Any(),
		types.NewOracleResultSignatureOrder(42, types.ENCODER_FULL_ABI),
		"",
		sdk.MustAccAddressFromBech32(request.Requester),
		request.FeeLimit,
	).DoAndReturn(func(
		ctx sdk.Context,
		content *types.OracleResultSignatureOrder,
		memo string,
		sender sdk.AccAddress,
		feeLimit sdk.Coins,
	) (bandtsstypes.SigningID, error) {
		ctx.KVStore(suite.key).Set([]byte{0xff, 0xff}, []byte("test"))
		return 0, tsstypes.ErrInsufficientSigners
	})

	k.ResolveSuccess(
		ctx,
		42,
		request.Requester,
		request.FeeLimit,
		basicResult,
		1234,
		request.TSSEncoder,
	)

	result := k.MustGetResult(ctx, 42)
	require.Equal(types.RESOLVE_STATUS_SUCCESS, result.ResolveStatus)
	require.Equal(basicResult, result.Result)
	require.Equal(sdk.Events{
		sdk.NewEvent(
			types.EventTypeHandleRequestSignFail,
			sdk.NewAttribute(types.AttributeKeyID, "42"),
			sdk.NewAttribute(types.AttributeKeyReason, "insufficient members for signing message"),
		),
		sdk.NewEvent(
			types.EventTypeResolve,
			sdk.NewAttribute(types.AttributeKeyID, "42"),
			sdk.NewAttribute(types.AttributeKeyResolveStatus, "1"),
			sdk.NewAttribute(types.AttributeKeyResult, "42415349435f524553554c54"), // BASIC_RESULT
			sdk.NewAttribute(types.AttributeKeyGasUsed, "1234"),
			sdk.NewAttribute(types.AttributeKeySigningErrCodespace, "tss"),
			sdk.NewAttribute(types.AttributeKeySigningErrCode, "22"),
		),
	}, ctx.EventManager().Events())

	// Check the signing request is saved correctly.
	require.Nil(ctx.KVStore(suite.key).Get([]byte{0xff, 0xff}))
	signingResult, err := k.GetSigningResult(ctx, 42)
	require.NoError(err)
	require.Equal(types.SigningResult{
		SigningID:      0,
		ErrorCodespace: "tss",
		ErrorCode:      22,
	}, signingResult)
}

func (suite *KeeperTestSuite) TestResolveSuccessAndGetSigningID() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()
	request := defaultRequest()
	request.TSSEncoder = types.ENCODER_FULL_ABI

	k.SetRequest(ctx, 42, request) // See report_test.go
	k.SetReport(ctx, 42, types.NewReport(validators[0].Address, true, nil))

	suite.bandtssKeeper.EXPECT().CreateDirectSigningRequest(
		gomock.Any(),
		types.NewOracleResultSignatureOrder(42, types.ENCODER_FULL_ABI),
		"",
		sdk.MustAccAddressFromBech32(request.Requester),
		request.FeeLimit,
	).DoAndReturn(func(
		ctx sdk.Context,
		content *types.OracleResultSignatureOrder,
		memo string,
		sender sdk.AccAddress,
		feeLimit sdk.Coins,
	) (bandtsstypes.SigningID, error) {
		ctx.KVStore(suite.key).Set([]byte{0xff, 0xff}, []byte("test"))
		return bandtsstypes.SigningID(1), nil
	})

	k.ResolveSuccess(
		ctx,
		42,
		request.Requester,
		request.FeeLimit,
		basicResult,
		1234,
		request.TSSEncoder,
	)

	result := k.MustGetResult(ctx, 42)
	require.Equal(types.RESOLVE_STATUS_SUCCESS, result.ResolveStatus)
	require.Equal(basicResult, result.Result)
	require.Equal(sdk.Events{
		sdk.NewEvent(
			types.EventTypeResolve,
			sdk.NewAttribute(types.AttributeKeyID, "42"),
			sdk.NewAttribute(types.AttributeKeyResolveStatus, "1"),
			sdk.NewAttribute(types.AttributeKeyResult, "42415349435f524553554c54"), // BASIC_RESULT
			sdk.NewAttribute(types.AttributeKeyGasUsed, "1234"),
			sdk.NewAttribute(types.AttributeKeySigningID, "1"),
		),
	}, ctx.EventManager().Events())

	// Check the signing request is saved correctly.
	require.Equal([]byte("test"), ctx.KVStore(suite.key).Get([]byte{0xff, 0xff}))
	signingResult, err := k.GetSigningResult(ctx, 42)
	require.NoError(err)
	require.Equal(types.SigningResult{SigningID: 1}, signingResult)
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
