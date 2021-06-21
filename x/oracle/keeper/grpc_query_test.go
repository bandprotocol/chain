package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/oracle/keeper"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

type RequestVerificationTestSuite struct {
	suite.Suite

	assert  *require.Assertions
	querier keeper.Querier
	request types.Request

	reporterPrivKey cryptotypes.PrivKey
	reporterAddr    sdk.AccAddress

	ctx sdk.Context
}

func (suite *RequestVerificationTestSuite) SetupTest() {
	suite.assert = require.New(suite.T())
	_, ctx, k := testapp.CreateTestInput(true)

	suite.querier = keeper.Querier{
		Keeper: k,
	}
	suite.ctx = ctx

	suite.request = types.NewRequest(
		1,
		BasicCalldata,
		[]sdk.ValAddress{testapp.Validators[0].ValAddress},
		1,
		1,
		testapp.ParseTime(0),
		"",
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("testdata")),
			types.NewRawRequest(2, 2, []byte("testdata")),
			types.NewRawRequest(3, 3, []byte("testdata")),
		},
		nil,
		0,
	)
	suite.reporterPrivKey = secp256k1.GenPrivKey()
	suite.reporterAddr = sdk.AccAddress(suite.reporterPrivKey.PubKey().Address())

	k.SetRequest(ctx, types.RequestID(1), suite.request)
	if err := k.AddReporter(ctx, testapp.Validators[0].ValAddress, suite.reporterAddr); err != nil {
		panic(err)
	}
}

func (suite *RequestVerificationTestSuite) TestSuccess() {
	req := &types.QueryRequestVerificationRequest{
		ChainId:    suite.ctx.ChainID(),
		Validator:  testapp.Validators[0].ValAddress.String(),
		RequestId:  1,
		ExternalId: 1,
		Reporter:   sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, suite.reporterPrivKey.PubKey()),
	}

	requestVerification := types.NewRequestVerification(req.ChainId, testapp.Validators[0].ValAddress, types.RequestID(req.RequestId), types.ExternalID(req.ExternalId))
	signature, err := suite.reporterPrivKey.Sign(requestVerification.GetSignBytes())
	suite.assert.NoError(err)
	req.Signature = signature

	res, err := suite.querier.RequestVerification(sdk.WrapSDKContext(suite.ctx), req)

	expectedResult := &types.QueryRequestVerificationResponse{
		ChainId:      suite.ctx.ChainID(),
		Validator:    testapp.Validators[0].ValAddress.String(),
		RequestId:    1,
		ExternalId:   1,
		DataSourceId: 1,
	}
	suite.assert.NoError(err, "RequestVerification should success")
	suite.assert.Equal(expectedResult, res, "Expected result should be matched")
}

func (suite *RequestVerificationTestSuite) TestFailedEmptyRequest() {
	res, err := suite.querier.RequestVerification(sdk.WrapSDKContext(suite.ctx), nil)

	suite.assert.Contains(err.Error(), "empty request", "RequestVerification should failed")
	suite.assert.Nil(res, "response should be nil")
}

func (suite *RequestVerificationTestSuite) TestFailedChainIDNotMatch() {
	req := &types.QueryRequestVerificationRequest{
		ChainId:    "other-chain-id",
		Validator:  testapp.Validators[0].ValAddress.String(),
		RequestId:  1,
		ExternalId: 1,
		Reporter:   sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, suite.reporterPrivKey.PubKey()),
	}

	requestVerification := types.NewRequestVerification(req.ChainId, testapp.Validators[0].ValAddress, types.RequestID(req.RequestId), types.ExternalID(req.ExternalId))
	signature, err := suite.reporterPrivKey.Sign(requestVerification.GetSignBytes())
	suite.assert.NoError(err)
	req.Signature = signature

	res, err := suite.querier.RequestVerification(sdk.WrapSDKContext(suite.ctx), req)

	suite.assert.Contains(err.Error(), "provided chain ID does not match the validator's chain ID", "RequestVerification should failed")
	suite.assert.Nil(res, "response should be nil")
}

func (suite *RequestVerificationTestSuite) TestFailedInvalidValidatorAddr() {
	req := &types.QueryRequestVerificationRequest{
		ChainId:    suite.ctx.ChainID(),
		Validator:  "someRandomString",
		RequestId:  1,
		ExternalId: 1,
		Reporter:   sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, suite.reporterPrivKey.PubKey()),
	}

	requestVerification := types.NewRequestVerification(req.ChainId, testapp.Validators[0].ValAddress, types.RequestID(req.RequestId), types.ExternalID(req.ExternalId))
	signature, err := suite.reporterPrivKey.Sign(requestVerification.GetSignBytes())
	suite.assert.NoError(err)
	req.Signature = signature

	res, err := suite.querier.RequestVerification(sdk.WrapSDKContext(suite.ctx), req)

	suite.assert.Contains(err.Error(), "unable to parse validator address", "RequestVerification should failed")
	suite.assert.Nil(res, "response should be nil")
}

func (suite *RequestVerificationTestSuite) TestFailedInvalidReporterPubKey() {
	req := &types.QueryRequestVerificationRequest{
		ChainId:    suite.ctx.ChainID(),
		Validator:  testapp.Validators[0].ValAddress.String(),
		RequestId:  1,
		ExternalId: 1,
		Reporter:   "someRandomString",
	}

	requestVerification := types.NewRequestVerification(req.ChainId, testapp.Validators[0].ValAddress, types.RequestID(req.RequestId), types.ExternalID(req.ExternalId))
	signature, err := suite.reporterPrivKey.Sign(requestVerification.GetSignBytes())
	suite.assert.NoError(err)
	req.Signature = signature

	res, err := suite.querier.RequestVerification(sdk.WrapSDKContext(suite.ctx), req)

	suite.assert.Contains(err.Error(), "unable to get reporter's public key", "RequestVerification should failed")
	suite.assert.Nil(res, "response should be nil")
}

func (suite *RequestVerificationTestSuite) TestFailedEmptySignature() {
	req := &types.QueryRequestVerificationRequest{
		ChainId:    suite.ctx.ChainID(),
		Validator:  testapp.Validators[0].ValAddress.String(),
		RequestId:  1,
		ExternalId: 1,
		Reporter:   sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, suite.reporterPrivKey.PubKey()),
	}

	res, err := suite.querier.RequestVerification(sdk.WrapSDKContext(suite.ctx), req)

	suite.assert.Contains(err.Error(), "invalid reporter's signature", "RequestVerification should failed")
	suite.assert.Nil(res, "response should be nil")
}

func (suite *RequestVerificationTestSuite) TestFailedReporterUnauthorized() {
	err := suite.querier.Keeper.RemoveReporter(suite.ctx, testapp.Validators[0].ValAddress, suite.reporterAddr)
	suite.assert.NoError(err)

	req := &types.QueryRequestVerificationRequest{
		ChainId:    suite.ctx.ChainID(),
		Validator:  testapp.Validators[0].ValAddress.String(),
		RequestId:  1,
		ExternalId: 1,
		Reporter:   sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, suite.reporterPrivKey.PubKey()),
	}

	requestVerification := types.NewRequestVerification(req.ChainId, testapp.Validators[0].ValAddress, types.RequestID(req.RequestId), types.ExternalID(req.ExternalId))
	signature, err := suite.reporterPrivKey.Sign(requestVerification.GetSignBytes())
	suite.assert.NoError(err)
	req.Signature = signature

	res, err := suite.querier.RequestVerification(sdk.WrapSDKContext(suite.ctx), req)

	suite.assert.Contains(err.Error(), "is not an authorized reporter of", "RequestVerification should failed")
	suite.assert.Nil(res, "response should be nil")
}

func (suite *RequestVerificationTestSuite) TestFailedUnselectedValidator() {
	suite.request.RequestedValidators = []string{testapp.Validators[1].ValAddress.String()}
	suite.querier.Keeper.SetRequest(suite.ctx, types.RequestID(1), suite.request)

	req := &types.QueryRequestVerificationRequest{
		ChainId:    suite.ctx.ChainID(),
		Validator:  testapp.Validators[0].ValAddress.String(),
		RequestId:  1,
		ExternalId: 1,
		Reporter:   sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, suite.reporterPrivKey.PubKey()),
	}

	requestVerification := types.NewRequestVerification(req.ChainId, testapp.Validators[0].ValAddress, types.RequestID(req.RequestId), types.ExternalID(req.ExternalId))
	signature, err := suite.reporterPrivKey.Sign(requestVerification.GetSignBytes())
	suite.assert.NoError(err)
	req.Signature = signature

	res, err := suite.querier.RequestVerification(sdk.WrapSDKContext(suite.ctx), req)

	suite.assert.Contains(err.Error(), "is not assigned for request ID", "RequestVerification should failed")
	suite.assert.Nil(res, "response should be nil")
}

func (suite *RequestVerificationTestSuite) TestFailedNoDataSourceFound() {
	suite.request.RawRequests = []types.RawRequest{}
	suite.querier.Keeper.SetRequest(suite.ctx, types.RequestID(1), suite.request)

	req := &types.QueryRequestVerificationRequest{
		ChainId:    suite.ctx.ChainID(),
		Validator:  testapp.Validators[0].ValAddress.String(),
		RequestId:  1,
		ExternalId: 1,
		Reporter:   sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, suite.reporterPrivKey.PubKey()),
	}

	requestVerification := types.NewRequestVerification(req.ChainId, testapp.Validators[0].ValAddress, types.RequestID(req.RequestId), types.ExternalID(req.ExternalId))
	signature, err := suite.reporterPrivKey.Sign(requestVerification.GetSignBytes())
	suite.assert.NoError(err)
	req.Signature = signature

	res, err := suite.querier.RequestVerification(sdk.WrapSDKContext(suite.ctx), req)

	suite.assert.Contains(err.Error(), "no data source required by the request", "RequestVerification should failed")
	suite.assert.Nil(res, "response should be nil")
}

func (suite *RequestVerificationTestSuite) TestFailedValidatorAlreadyReported() {
	err := suite.querier.Keeper.AddReport(suite.ctx, types.RequestID(1), types.NewReport(testapp.Validators[0].ValAddress, true, []types.RawReport{
		types.NewRawReport(1, 0, []byte("testdata")),
		types.NewRawReport(2, 0, []byte("testdata")),
		types.NewRawReport(3, 0, []byte("testdata")),
	}))
	suite.assert.NoError(err)

	req := &types.QueryRequestVerificationRequest{
		ChainId:    suite.ctx.ChainID(),
		Validator:  testapp.Validators[0].ValAddress.String(),
		RequestId:  1,
		ExternalId: 1,
		Reporter:   sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, suite.reporterPrivKey.PubKey()),
	}

	requestVerification := types.NewRequestVerification(req.ChainId, testapp.Validators[0].ValAddress, types.RequestID(req.RequestId), types.ExternalID(req.ExternalId))
	signature, err := suite.reporterPrivKey.Sign(requestVerification.GetSignBytes())
	suite.assert.NoError(err)
	req.Signature = signature

	res, err := suite.querier.RequestVerification(sdk.WrapSDKContext(suite.ctx), req)

	suite.assert.Contains(err.Error(), "already submitted data report", "RequestVerification should failed")
	suite.assert.Nil(res, "response should be nil")
}

func (suite *RequestVerificationTestSuite) TestFailedRequestAlreadyExpired() {
	req := &types.QueryRequestVerificationRequest{
		ChainId:    suite.ctx.ChainID(),
		Validator:  testapp.Validators[0].ValAddress.String(),
		RequestId:  1,
		ExternalId: 1,
		Reporter:   sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, suite.reporterPrivKey.PubKey()),
	}

	suite.ctx = suite.ctx.WithBlockHeight(1000)

	requestVerification := types.NewRequestVerification(req.ChainId, testapp.Validators[0].ValAddress, types.RequestID(req.RequestId), types.ExternalID(req.ExternalId))
	signature, err := suite.reporterPrivKey.Sign(requestVerification.GetSignBytes())
	suite.assert.NoError(err)
	req.Signature = signature

	res, err := suite.querier.RequestVerification(sdk.WrapSDKContext(suite.ctx), req)

	suite.assert.Contains(err.Error(), "Request with ID 1 is already expired", "RequestVerification should failed")
	suite.assert.Nil(res, "response should be nil")
}

type PendingRequestsTestSuite struct {
	suite.Suite

	assert  *require.Assertions
	querier keeper.Querier

	ctx sdk.Context
}

func (suite *PendingRequestsTestSuite) SetupTest() {
	suite.assert = require.New(suite.T())
	_, ctx, k := testapp.CreateTestInput(true)

	suite.querier = keeper.Querier{
		Keeper: k,
	}
	suite.ctx = ctx
}

func (suite *PendingRequestsTestSuite) TestSuccess() {
	assignedButPendingReq := types.NewRequest(
		1,
		BasicCalldata,
		[]sdk.ValAddress{testapp.Validators[0].ValAddress},
		1,
		1,
		testapp.ParseTime(0),
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
		BasicCalldata,
		[]sdk.ValAddress{testapp.Validators[1].ValAddress},
		1,
		1,
		testapp.ParseTime(0),
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
		BasicCalldata,
		[]sdk.ValAddress{
			testapp.Validators[0].ValAddress,
			testapp.Validators[1].ValAddress,
		},
		1,
		1,
		testapp.ParseTime(0),
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
		BasicCalldata,
		[]sdk.ValAddress{
			testapp.Validators[0].ValAddress,
			testapp.Validators[1].ValAddress,
		},
		1,
		1,
		testapp.ParseTime(0),
		"",
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("testdata")),
			types.NewRawRequest(2, 2, []byte("testdata")),
			types.NewRawRequest(3, 3, []byte("testdata")),
		},
		nil,
		0,
	)

	suite.querier.Keeper.SetRequest(suite.ctx, types.RequestID(3), assignedButPendingReq)
	suite.querier.Keeper.SetRequest(suite.ctx, types.RequestID(4), notBeAssignedReq)
	suite.querier.Keeper.SetRequest(suite.ctx, types.RequestID(5), alreadyReportAllReq)
	suite.querier.Keeper.SetRequest(suite.ctx, types.RequestID(6), assignedButReportedReq)
	suite.querier.Keeper.SetRequestCount(suite.ctx, 4)
	suite.querier.Keeper.SetRequestLastExpired(suite.ctx, 2)
	suite.querier.Keeper.SetReport(suite.ctx, 5, types.NewReport(testapp.Validators[0].ValAddress, true, []types.RawReport{
		types.NewRawReport(1, 0, []byte("testdata")),
		types.NewRawReport(2, 0, []byte("testdata")),
		types.NewRawReport(3, 0, []byte("testdata")),
	}))
	suite.querier.Keeper.SetReport(suite.ctx, 5, types.NewReport(testapp.Validators[1].ValAddress, true, []types.RawReport{
		types.NewRawReport(1, 0, []byte("testdata")),
		types.NewRawReport(2, 0, []byte("testdata")),
		types.NewRawReport(3, 0, []byte("testdata")),
	}))
	suite.querier.Keeper.SetReport(suite.ctx, 6, types.NewReport(testapp.Validators[0].ValAddress, true, []types.RawReport{
		types.NewRawReport(1, 0, []byte("testdata")),
		types.NewRawReport(2, 0, []byte("testdata")),
		types.NewRawReport(3, 0, []byte("testdata")),
	}))

	r, err := suite.querier.PendingRequests(sdk.WrapSDKContext(suite.ctx), &types.QueryPendingRequestsRequest{
		ValidatorAddress: sdk.ValAddress(testapp.Validators[0].Address).String(),
	})

	suite.assert.Equal(&types.QueryPendingRequestsResponse{RequestIDs: []int64{3}}, r)
	suite.assert.NoError(err)
}

func TestRequestVerification(t *testing.T) {
	suite.Run(t, new(RequestVerificationTestSuite))
}

func TestPendingRequests(t *testing.T) {
	suite.Run(t, new(PendingRequestsTestSuite))
}
