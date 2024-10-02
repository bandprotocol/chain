package keeper_test

import (
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

func (suite *KeeperTestSuite) TestHasReport() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// We should not have a report to request ID 42 from Alice without setting it.
	require.False(k.HasReport(ctx, 42, bandtesting.Alice.ValAddress))
	// After we set it, we should be able to find it.
	k.SetReport(ctx, 42, types.NewReport(bandtesting.Alice.ValAddress, true, nil))
	require.True(k.HasReport(ctx, 42, bandtesting.Alice.ValAddress))
}

func (suite *KeeperTestSuite) TestAddReportSuccess() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	k.SetRequest(ctx, 1, defaultRequest())
	err := k.AddReport(ctx, 1,
		validators[0].Address, true, []types.RawReport{
			types.NewRawReport(1, 0, []byte("data1/1")),
			types.NewRawReport(2, 1, []byte("data2/1")),
			types.NewRawReport(3, 0, []byte("data3/1")),
		},
	)
	require.NoError(err)
	require.Equal([]types.Report{
		types.NewReport(validators[0].Address, true, []types.RawReport{
			types.NewRawReport(1, 0, []byte("data1/1")),
			types.NewRawReport(2, 1, []byte("data2/1")),
			types.NewRawReport(3, 0, []byte("data3/1")),
		}),
	}, k.GetReports(ctx, 1))
}

func (suite *KeeperTestSuite) TestReportOnNonExistingRequest() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	err := k.AddReport(ctx, 1,
		validators[0].Address, true, []types.RawReport{
			types.NewRawReport(42, 0, []byte("data1/1")),
			types.NewRawReport(43, 1, []byte("data2/1")),
		},
	)
	require.ErrorIs(err, types.ErrRequestNotFound)
}

func (suite *KeeperTestSuite) TestReportByNotRequestedValidator() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	k.SetRequest(ctx, 1, defaultRequest())
	err := k.AddReport(ctx, 1,
		bandtesting.Alice.ValAddress, true, []types.RawReport{
			types.NewRawReport(42, 0, []byte("data1/1")),
			types.NewRawReport(43, 1, []byte("data2/1")),
		},
	)
	require.ErrorIs(err, types.ErrValidatorNotRequested)
}

func (suite *KeeperTestSuite) TestDuplicateReport() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	k.SetRequest(ctx, 1, defaultRequest())
	err := k.AddReport(ctx, 1,
		validators[0].Address, true, []types.RawReport{
			types.NewRawReport(1, 0, []byte("data1/1")),
			types.NewRawReport(2, 1, []byte("data2/1")),
			types.NewRawReport(3, 0, []byte("data3/1")),
		},
	)
	require.NoError(err)
	err = k.AddReport(ctx, 1,
		validators[0].Address, true, []types.RawReport{
			types.NewRawReport(1, 0, []byte("data1/1")),
			types.NewRawReport(2, 1, []byte("data2/1")),
			types.NewRawReport(3, 0, []byte("data3/1")),
		},
	)
	require.ErrorIs(err, types.ErrValidatorAlreadyReported)
}

func (suite *KeeperTestSuite) TestReportInvalidDataSourceCount() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	k.SetRequest(ctx, 1, defaultRequest())
	err := k.AddReport(ctx, 1,
		validators[0].Address, true, []types.RawReport{
			types.NewRawReport(42, 0, []byte("data1/1")),
		},
	)
	require.ErrorIs(err, types.ErrInvalidReportSize)
}

func (suite *KeeperTestSuite) TestReportInvalidExternalIDs() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	k.SetRequest(ctx, 1, defaultRequest())
	err := k.AddReport(ctx, 1,
		validators[0].Address, true, []types.RawReport{
			types.NewRawReport(1, 0, []byte("data1/1")),
			types.NewRawReport(44, 1, []byte("data2/1")), // Bad External ID
			types.NewRawReport(3, 0, []byte("data3/1")),
		},
	)
	require.ErrorIs(err, types.ErrRawRequestNotFound)
}

func (suite *KeeperTestSuite) TestGetReportCount() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// We start by setting some arbitrary reports.
	k.SetReport(ctx, types.RequestID(1), types.NewReport(bandtesting.Alice.ValAddress, true, []types.RawReport{}))
	k.SetReport(ctx, types.RequestID(1), types.NewReport(bandtesting.Bob.ValAddress, true, []types.RawReport{}))
	k.SetReport(ctx, types.RequestID(2), types.NewReport(bandtesting.Alice.ValAddress, true, []types.RawReport{}))
	k.SetReport(ctx, types.RequestID(2), types.NewReport(bandtesting.Bob.ValAddress, true, []types.RawReport{}))
	k.SetReport(ctx, types.RequestID(2), types.NewReport(bandtesting.Carol.ValAddress, true, []types.RawReport{}))
	// GetReportCount should return the correct values.
	require.Equal(uint64(2), k.GetReportCount(ctx, types.RequestID(1)))
	require.Equal(uint64(3), k.GetReportCount(ctx, types.RequestID(2)))
}

func (suite *KeeperTestSuite) TestDeleteReports() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// We start by setting some arbitrary reports.
	k.SetReport(ctx, types.RequestID(1), types.NewReport(bandtesting.Alice.ValAddress, true, []types.RawReport{}))
	k.SetReport(ctx, types.RequestID(1), types.NewReport(bandtesting.Bob.ValAddress, true, []types.RawReport{}))
	k.SetReport(ctx, types.RequestID(2), types.NewReport(bandtesting.Alice.ValAddress, true, []types.RawReport{}))
	k.SetReport(ctx, types.RequestID(2), types.NewReport(bandtesting.Bob.ValAddress, true, []types.RawReport{}))
	k.SetReport(ctx, types.RequestID(2), types.NewReport(bandtesting.Carol.ValAddress, true, []types.RawReport{}))
	// All reports should exist on the state.
	require.True(k.HasReport(ctx, types.RequestID(1), bandtesting.Alice.ValAddress))
	require.True(k.HasReport(ctx, types.RequestID(1), bandtesting.Bob.ValAddress))
	require.True(k.HasReport(ctx, types.RequestID(2), bandtesting.Alice.ValAddress))
	require.True(k.HasReport(ctx, types.RequestID(2), bandtesting.Bob.ValAddress))
	require.True(k.HasReport(ctx, types.RequestID(2), bandtesting.Carol.ValAddress))
	// After we delete reports related to request#1, they must disappear.
	k.DeleteReports(ctx, types.RequestID(1))
	require.False(k.HasReport(ctx, types.RequestID(1), bandtesting.Alice.ValAddress))
	require.False(k.HasReport(ctx, types.RequestID(1), bandtesting.Bob.ValAddress))
	require.True(k.HasReport(ctx, types.RequestID(2), bandtesting.Alice.ValAddress))
	require.True(k.HasReport(ctx, types.RequestID(2), bandtesting.Bob.ValAddress))
	require.True(k.HasReport(ctx, types.RequestID(2), bandtesting.Carol.ValAddress))
}
