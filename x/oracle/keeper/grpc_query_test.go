package keeper_test

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/oracle/keeper"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

// -----------------------------------------
// --- Test for QueryRequestVerification ---
// -----------------------------------------

func (suite *KeeperTestSuite) TestRequestVerificationValid() {
	ctx := suite.ctx
	querier := suite.queryClient
	require := suite.Require()
	k := suite.oracleKeeper
	k.SetRequest(ctx, types.RequestID(1), defaultRequest())
	k.SetRequestCount(ctx, 1)

	req := &types.QueryRequestVerificationRequest{
		ChainId:      suite.ctx.ChainID(),
		Validator:    validators[0].Address.String(),
		RequestId:    1,
		ExternalId:   1,
		DataSourceId: 1,
		Reporter:     hex.EncodeToString(reporterPubKey.Bytes()),
	}

	requestVerification := types.NewRequestVerification(
		req.ChainId,
		validators[0].Address,
		types.RequestID(req.RequestId),
		types.ExternalID(req.ExternalId),
		types.DataSourceID(req.DataSourceId),
	)
	signature, err := reporterPrivKey.Sign(requestVerification.GetSignBytes())
	require.NoError(err)
	req.Signature = signature

	res, err := querier.RequestVerification(context.Background(), req)

	expectedResult := &types.QueryRequestVerificationResponse{
		ChainId:      ctx.ChainID(),
		Validator:    validators[0].Address.String(),
		RequestId:    1,
		ExternalId:   1,
		DataSourceId: 1,
		IsDelay:      false,
	}
	require.NoError(err, "RequestVerification should success")
	require.Equal(expectedResult, res, "Expected result should be matched")
}

func (suite *KeeperTestSuite) TestRequestVerificationFailedRequestIDNotExist() {
	querier := suite.queryClient
	require := suite.Require()

	req := &types.QueryRequestVerificationRequest{
		ChainId:      suite.ctx.ChainID(),
		Validator:    validators[0].Address.String(),
		RequestId:    2,
		ExternalId:   1,
		DataSourceId: 1,
		Reporter:     hex.EncodeToString(reporterPrivKey.PubKey().Bytes()),
	}

	requestVerification := types.NewRequestVerification(
		req.ChainId,
		validators[0].Address,
		types.RequestID(req.RequestId),
		types.ExternalID(req.ExternalId),
		types.DataSourceID(req.DataSourceId),
	)
	signature, err := reporterPrivKey.Sign(requestVerification.GetSignBytes())
	require.NoError(err)
	req.Signature = signature

	res, err := querier.RequestVerification(context.Background(), req)

	require.Contains(err.Error(), "unable to get request from chain", "RequestVerification should failed")
	require.Nil(res, "response should be nil")
}

func (suite *KeeperTestSuite) TestRequestVerificationInDelayRange() {
	ctx := suite.ctx
	querier := suite.queryClient
	require := suite.Require()

	req := &types.QueryRequestVerificationRequest{
		ChainId:      suite.ctx.ChainID(),
		Validator:    validators[0].Address.String(),
		RequestId:    5,
		ExternalId:   1,
		DataSourceId: 1,
		Reporter:     hex.EncodeToString(reporterPrivKey.PubKey().Bytes()),
		MaxDelay:     5,
	}

	requestVerification := types.NewRequestVerification(
		req.ChainId,
		validators[0].Address,
		types.RequestID(req.RequestId),
		types.ExternalID(req.ExternalId),
		types.DataSourceID(req.DataSourceId),
	)
	signature, err := reporterPrivKey.Sign(requestVerification.GetSignBytes())
	require.NoError(err)
	req.Signature = signature

	res, err := querier.RequestVerification(context.Background(), req)

	expectedResult := &types.QueryRequestVerificationResponse{
		ChainId:      ctx.ChainID(),
		Validator:    validators[0].Address.String(),
		RequestId:    5,
		ExternalId:   1,
		DataSourceId: 1,
		IsDelay:      true,
	}
	require.NoError(err, "RequestVerification should success")
	require.Equal(expectedResult, res, "Expected result should be matched")
}

func (suite *KeeperTestSuite) TestRequestVerificationFailedExceedDelayRange() {
	ctx := suite.ctx
	querier := suite.queryClient
	require := suite.Require()

	req := &types.QueryRequestVerificationRequest{
		ChainId:      ctx.ChainID(),
		Validator:    validators[0].Address.String(),
		RequestId:    6,
		ExternalId:   1,
		DataSourceId: 1,
		Reporter:     hex.EncodeToString(reporterPrivKey.PubKey().Bytes()),
		MaxDelay:     5,
	}

	requestVerification := types.NewRequestVerification(
		req.ChainId,
		validators[0].Address,
		types.RequestID(req.RequestId),
		types.ExternalID(req.ExternalId),
		types.DataSourceID(req.DataSourceId),
	)
	signature, err := reporterPrivKey.Sign(requestVerification.GetSignBytes())
	require.NoError(err)
	req.Signature = signature

	res, err := querier.RequestVerification(context.Background(), req)

	require.Contains(err.Error(), "unable to get request from chain", "RequestVerification should failed")
	require.Nil(res, "response should be nil")
}

func (suite *KeeperTestSuite) TestRequestVerificationFailedDataSourceIDNotMatch() {
	ctx := suite.ctx
	querier := suite.queryClient
	require := suite.Require()
	k := suite.oracleKeeper
	k.SetRequest(ctx, types.RequestID(1), defaultRequest())
	k.SetRequestCount(ctx, 1)

	req := &types.QueryRequestVerificationRequest{
		ChainId:      ctx.ChainID(),
		Validator:    validators[0].Address.String(),
		RequestId:    1,
		ExternalId:   1,
		DataSourceId: 2,
		Reporter:     hex.EncodeToString(reporterPrivKey.PubKey().Bytes()),
	}

	requestVerification := types.NewRequestVerification(
		req.ChainId,
		validators[0].Address,
		types.RequestID(req.RequestId),
		types.ExternalID(req.ExternalId),
		types.DataSourceID(req.DataSourceId),
	)
	signature, err := reporterPrivKey.Sign(requestVerification.GetSignBytes())
	require.NoError(err)
	req.Signature = signature

	res, err := querier.RequestVerification(context.Background(), req)

	require.Contains(
		err.Error(),
		"is not match with data source id provided in request",
		"RequestVerification should failed",
	)
	require.Nil(res, "response should be nil")
}

func (suite *KeeperTestSuite) TestRequestVerificationFailedChainIDNotMatch() {
	querier := suite.queryClient
	require := suite.Require()

	req := &types.QueryRequestVerificationRequest{
		ChainId:      "other-chain-id",
		Validator:    validators[0].Address.String(),
		RequestId:    1,
		ExternalId:   1,
		DataSourceId: 1,
		Reporter:     hex.EncodeToString(reporterPrivKey.PubKey().Bytes()),
	}

	requestVerification := types.NewRequestVerification(
		req.ChainId,
		validators[0].Address,
		types.RequestID(req.RequestId),
		types.ExternalID(req.ExternalId),
		types.DataSourceID(req.DataSourceId),
	)
	signature, err := reporterPrivKey.Sign(requestVerification.GetSignBytes())
	require.NoError(err)
	req.Signature = signature

	res, err := querier.RequestVerification(context.Background(), req)

	require.Contains(
		err.Error(),
		"provided chain ID does not match the validator's chain ID",
		"RequestVerification should failed",
	)
	require.Nil(res, "response should be nil")
}

func (suite *KeeperTestSuite) TestRequestVerificationFailedInvalidValidatorAddr() {
	ctx := suite.ctx
	querier := suite.queryClient
	require := suite.Require()

	req := &types.QueryRequestVerificationRequest{
		ChainId:      ctx.ChainID(),
		Validator:    "someRandomString",
		RequestId:    1,
		ExternalId:   1,
		DataSourceId: 1,
		Reporter:     hex.EncodeToString(reporterPrivKey.PubKey().Bytes()),
	}

	requestVerification := types.NewRequestVerification(
		req.ChainId,
		validators[0].Address,
		types.RequestID(req.RequestId),
		types.ExternalID(req.ExternalId),
		types.DataSourceID(req.DataSourceId),
	)
	signature, err := reporterPrivKey.Sign(requestVerification.GetSignBytes())
	require.NoError(err)
	req.Signature = signature

	res, err := querier.RequestVerification(context.Background(), req)

	require.Contains(err.Error(), "unable to parse validator address", "RequestVerification should failed")
	require.Nil(res, "response should be nil")
}

func (suite *KeeperTestSuite) TestRequestVerificationFailedInvalidReporterPubKey() {
	ctx := suite.ctx
	querier := suite.queryClient
	require := suite.Require()

	req := &types.QueryRequestVerificationRequest{
		ChainId:      ctx.ChainID(),
		Validator:    validators[0].Address.String(),
		RequestId:    1,
		ExternalId:   1,
		DataSourceId: 1,
		Reporter:     "RANDOM STRING",
	}

	requestVerification := types.NewRequestVerification(
		req.ChainId,
		validators[0].Address,
		types.RequestID(req.RequestId),
		types.ExternalID(req.ExternalId),
		types.DataSourceID(req.DataSourceId),
	)
	signature, err := reporterPrivKey.Sign(requestVerification.GetSignBytes())
	require.NoError(err)
	req.Signature = signature

	res, err := querier.RequestVerification(context.Background(), req)

	require.Contains(err.Error(), "unable to get reporter's public key", "RequestVerification should failed")
	require.Nil(res, "response should be nil")
}

func (suite *KeeperTestSuite) TestRequestVerificationFailedEmptySignature() {
	ctx := suite.ctx
	querier := suite.queryClient
	require := suite.Require()

	req := &types.QueryRequestVerificationRequest{
		ChainId:    ctx.ChainID(),
		Validator:  validators[0].Address.String(),
		RequestId:  1,
		ExternalId: 1,
		Reporter:   hex.EncodeToString(reporterPrivKey.PubKey().Bytes()),
	}

	res, err := querier.RequestVerification(context.Background(), req)

	require.Contains(err.Error(), "invalid reporter's signature", "RequestVerification should failed")
	require.Nil(res, "response should be nil")
}

func (suite *KeeperTestSuite) TestRequestVerificationFailedReporterUnauthorized() {
	ctx := suite.ctx
	querier := suite.queryClient
	require := suite.Require()

	req := &types.QueryRequestVerificationRequest{
		ChainId:      ctx.ChainID(),
		Validator:    validators[1].Address.String(),
		RequestId:    1,
		ExternalId:   1,
		DataSourceId: 1,
		Reporter:     hex.EncodeToString(reporterPrivKey.PubKey().Bytes()),
	}

	requestVerification := types.NewRequestVerification(
		req.ChainId,
		validators[1].Address,
		types.RequestID(req.RequestId),
		types.ExternalID(req.ExternalId),
		types.DataSourceID(req.DataSourceId),
	)
	signature, err := reporterPrivKey.Sign(requestVerification.GetSignBytes())
	require.NoError(err)
	req.Signature = signature

	res, err := querier.RequestVerification(context.Background(), req)

	require.Contains(err.Error(), "is not an authorized reporter of", "RequestVerification should failed")
	require.Nil(res, "response should be nil")
}

func (suite *KeeperTestSuite) TestRequestVerificationFailedUnselectedValidator() {
	ctx := suite.ctx
	querier := suite.queryClient
	require := suite.Require()
	k := suite.oracleKeeper

	request := defaultRequest()
	request.RequestedValidators = []string{validators[1].Address.String()}

	k.SetRequest(ctx, types.RequestID(1), request)
	k.SetRequestCount(ctx, 1)

	req := &types.QueryRequestVerificationRequest{
		ChainId:      ctx.ChainID(),
		Validator:    validators[0].Address.String(),
		RequestId:    1,
		ExternalId:   1,
		DataSourceId: 1,
		Reporter:     hex.EncodeToString(reporterPrivKey.PubKey().Bytes()),
	}

	requestVerification := types.NewRequestVerification(
		req.ChainId,
		validators[0].Address,
		types.RequestID(req.RequestId),
		types.ExternalID(req.ExternalId),
		types.DataSourceID(req.DataSourceId),
	)
	signature, err := reporterPrivKey.Sign(requestVerification.GetSignBytes())
	require.NoError(err)
	req.Signature = signature

	res, err := querier.RequestVerification(context.Background(), req)

	require.Contains(err.Error(), "is not assigned for request ID", "RequestVerification should failed")
	require.Nil(res, "response should be nil")
}

func (suite *KeeperTestSuite) TestRequestVerificationFailedNoDataSourceFound() {
	ctx := suite.ctx
	querier := suite.queryClient
	require := suite.Require()
	k := suite.oracleKeeper

	request := defaultRequest()
	request.RawRequests = []types.RawRequest{}
	k.SetRequest(ctx, types.RequestID(1), request)

	req := &types.QueryRequestVerificationRequest{
		ChainId:      ctx.ChainID(),
		Validator:    validators[0].Address.String(),
		RequestId:    1,
		ExternalId:   1,
		DataSourceId: 1,
		Reporter:     hex.EncodeToString(reporterPrivKey.PubKey().Bytes()),
	}

	requestVerification := types.NewRequestVerification(
		req.ChainId,
		validators[0].Address,
		types.RequestID(req.RequestId),
		types.ExternalID(req.ExternalId),
		types.DataSourceID(req.DataSourceId),
	)
	signature, err := reporterPrivKey.Sign(requestVerification.GetSignBytes())
	require.NoError(err)
	req.Signature = signature

	res, err := querier.RequestVerification(context.Background(), req)

	require.Contains(err.Error(), "no data source required by the request", "RequestVerification should failed")
	require.Nil(res, "response should be nil")
}

func (suite *KeeperTestSuite) TestRequestVerificationFailedValidatorAlreadyReported() {
	ctx := suite.ctx
	querier := suite.queryClient
	require := suite.Require()
	k := suite.oracleKeeper
	k.SetRequest(ctx, types.RequestID(1), defaultRequest())
	k.SetRequestCount(ctx, 1)

	err := k.AddReport(
		ctx,
		types.RequestID(1),
		validators[0].Address, true, []types.RawReport{
			types.NewRawReport(1, 0, []byte("testdata")),
			types.NewRawReport(2, 0, []byte("testdata")),
			types.NewRawReport(3, 0, []byte("testdata")),
		},
	)
	require.NoError(err)

	req := &types.QueryRequestVerificationRequest{
		ChainId:      ctx.ChainID(),
		Validator:    validators[0].Address.String(),
		RequestId:    1,
		ExternalId:   1,
		DataSourceId: 1,
		Reporter:     hex.EncodeToString(reporterPrivKey.PubKey().Bytes()),
	}

	requestVerification := types.NewRequestVerification(
		req.ChainId,
		validators[0].Address,
		types.RequestID(req.RequestId),
		types.ExternalID(req.ExternalId),
		types.DataSourceID(req.DataSourceId),
	)
	signature, err := reporterPrivKey.Sign(requestVerification.GetSignBytes())
	require.NoError(err)
	req.Signature = signature

	res, err := querier.RequestVerification(context.Background(), req)

	require.Contains(err.Error(), "already submitted data report", "RequestVerification should failed")
	require.Nil(res, "response should be nil")
}

func (suite *KeeperTestSuite) TestRequestVerificationFailedRequestAlreadyExpired() {
	ctx := suite.ctx
	require := suite.Require()
	k := suite.oracleKeeper
	k.SetRequest(ctx, types.RequestID(1), defaultRequest())
	k.SetRequestCount(ctx, 1)

	ctx = ctx.WithBlockHeight(1000)
	encCfg := moduletestutil.MakeTestEncodingConfig()
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, keeper.Querier{
		Keeper: suite.oracleKeeper,
	})
	querier := types.NewQueryClient(queryHelper)

	req := &types.QueryRequestVerificationRequest{
		ChainId:      ctx.ChainID(),
		Validator:    validators[0].Address.String(),
		RequestId:    1,
		ExternalId:   1,
		DataSourceId: 1,
		Reporter:     hex.EncodeToString(reporterPrivKey.PubKey().Bytes()),
	}

	requestVerification := types.NewRequestVerification(
		req.ChainId,
		validators[0].Address,
		types.RequestID(req.RequestId),
		types.ExternalID(req.ExternalId),
		types.DataSourceID(req.DataSourceId),
	)
	signature, err := reporterPrivKey.Sign(requestVerification.GetSignBytes())
	require.NoError(err)
	req.Signature = signature

	res, err := querier.RequestVerification(ctx, req)

	require.Contains(err.Error(), "Request with ID 1 is already expired", "RequestVerification should failed")
	require.Nil(res, "response should be nil")
}

// ------------------------------
// --- Test for QueryReporters --
// ------------------------------

func (suite *KeeperTestSuite) TestGetReporters() {
	querier := suite.queryClient
	require := suite.Require()

	req := &types.QueryReportersRequest{
		ValidatorAddress: validators[0].Address.String(),
	}
	res, err := querier.Reporters(context.Background(), req)

	expectedResult := &types.QueryReportersResponse{
		Reporter: []string{reporterAddr.String()},
	}
	require.NoError(err, "Reporters should success")
	require.Equal(expectedResult, res, "Expected result should be matched")
}

func (suite *KeeperTestSuite) TestGetExpiredReporters() {
	ctx := suite.ctx
	require := suite.Require()

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(10 * time.Minute))
	encCfg := moduletestutil.MakeTestEncodingConfig()
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, keeper.Querier{
		Keeper: suite.oracleKeeper,
	})
	querier := types.NewQueryClient(queryHelper)

	req := &types.QueryReportersRequest{
		ValidatorAddress: validators[0].Address.String(),
	}
	res, err := querier.Reporters(ctx, req)

	expectedResult := &types.QueryReportersResponse{
		Reporter: []string(nil),
	}
	require.NoError(err, "Reporters should success")
	require.Equal(expectedResult, res, "Expected result should be matched")
}

// -------------------------------
// --- Test for QueryIsReporter --
// -------------------------------

func (suite *KeeperTestSuite) TestIsReporter() {
	querier := suite.queryClient
	require := suite.Require()

	req := &types.QueryIsReporterRequest{
		ValidatorAddress: validators[0].Address.String(),
		ReporterAddress:  reporterAddr.String(),
	}
	res, err := querier.IsReporter(context.Background(), req)

	expectedResult := &types.QueryIsReporterResponse{
		IsReporter: true,
	}
	require.NoError(err, "IsReporter should success")
	require.Equal(expectedResult, res, "Expected result should be matched")
}

func (suite *KeeperTestSuite) TestIsNotReporter() {
	querier := suite.queryClient
	require := suite.Require()

	req := &types.QueryIsReporterRequest{
		ValidatorAddress: validators[1].Address.String(),
		ReporterAddress:  reporterAddr.String(),
	}
	res, err := querier.IsReporter(context.Background(), req)

	expectedResult := &types.QueryIsReporterResponse{
		IsReporter: false,
	}
	require.NoError(err, "IsReporter should success")
	require.Equal(expectedResult, res, "Expected result should be matched")
}

// ------------------------------------
// --- Test for QueryPendingRequests --
// ------------------------------------

func (suite *KeeperTestSuite) TestPendingRequestsSuccess() {
	ctx := suite.ctx
	querier := suite.queryClient
	require := suite.Require()
	k := suite.oracleKeeper

	assignedButPendingReq := types.NewRequest(
		1,
		basicCalldata,
		[]sdk.ValAddress{validators[0].Address},
		1,
		1,
		bandtesting.ParseTime(0),
		"",
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("testdata")),
			types.NewRawRequest(2, 2, []byte("testdata")),
			types.NewRawRequest(3, 3, []byte("testdata")),
		},
		nil,
		0,
	)
	notBeAssignedReq := types.NewRequest(
		1,
		basicCalldata,
		[]sdk.ValAddress{validators[1].Address},
		1,
		1,
		bandtesting.ParseTime(0),
		"",
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("testdata")),
			types.NewRawRequest(2, 2, []byte("testdata")),
			types.NewRawRequest(3, 3, []byte("testdata")),
		},
		nil,
		0,
	)
	alreadyReportAllReq := types.NewRequest(
		1,
		basicCalldata,
		[]sdk.ValAddress{
			validators[0].Address,
			validators[1].Address,
		},
		1,
		1,
		bandtesting.ParseTime(0),
		"",
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("testdata")),
			types.NewRawRequest(2, 2, []byte("testdata")),
			types.NewRawRequest(3, 3, []byte("testdata")),
		},
		nil,
		0,
	)
	assignedButReportedReq := types.NewRequest(
		1,
		basicCalldata,
		[]sdk.ValAddress{
			validators[0].Address,
			validators[1].Address,
		},
		1,
		1,
		bandtesting.ParseTime(0),
		"",
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("testdata")),
			types.NewRawRequest(2, 2, []byte("testdata")),
			types.NewRawRequest(3, 3, []byte("testdata")),
		},
		nil,
		0,
	)

	k.SetRequest(ctx, types.RequestID(3), assignedButPendingReq)
	k.SetRequest(ctx, types.RequestID(4), notBeAssignedReq)
	k.SetRequest(ctx, types.RequestID(5), alreadyReportAllReq)
	k.SetRequest(ctx, types.RequestID(6), assignedButReportedReq)
	k.SetRequestCount(ctx, 4)
	k.SetRequestLastExpired(ctx, 2)
	k.SetReport(
		ctx,
		5,
		types.NewReport(validators[0].Address, true, []types.RawReport{
			types.NewRawReport(1, 0, []byte("testdata")),
			types.NewRawReport(2, 0, []byte("testdata")),
			types.NewRawReport(3, 0, []byte("testdata")),
		}),
	)
	k.SetReport(
		ctx,
		5,
		types.NewReport(validators[1].Address, true, []types.RawReport{
			types.NewRawReport(1, 0, []byte("testdata")),
			types.NewRawReport(2, 0, []byte("testdata")),
			types.NewRawReport(3, 0, []byte("testdata")),
		}),
	)
	k.SetReport(
		ctx,
		6,
		types.NewReport(validators[0].Address, true, []types.RawReport{
			types.NewRawReport(1, 0, []byte("testdata")),
			types.NewRawReport(2, 0, []byte("testdata")),
			types.NewRawReport(3, 0, []byte("testdata")),
		}),
	)

	r, err := querier.PendingRequests(context.Background(), &types.QueryPendingRequestsRequest{
		ValidatorAddress: sdk.ValAddress(sdk.AccAddress(validators[0].Address)).String(),
	})

	require.Equal(&types.QueryPendingRequestsResponse{RequestIDs: []uint64{3}}, r)
	require.NoError(err)
}
