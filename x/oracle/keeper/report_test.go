package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

// TODO: Fix tests
// import (
// 	"testing"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/stretchr/testify/require"

// 	bandtesting "github.com/bandprotocol/chain/v3/testing"
// 	"github.com/bandprotocol/chain/v3/x/oracle/types"
// )

func defaultRequest() types.Request {
	return types.NewRequest(
		1, basicCalldata,
		[]sdk.ValAddress{validators[0].Address, validators[1].Address},
		2, 0, bandtesting.ParseTime(0),
		basicClientID, []types.RawRequest{
			types.NewRawRequest(42, 1, basicCalldata),
			types.NewRawRequest(43, 2, basicCalldata),
		}, nil, 0,
	)
}

// func TestHasReport(t *testing.T) {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	// We should not have a report to request ID 42 from Alice without setting it.
// 	require.False(t, k.HasReport(ctx, 42, bandtesting.Alice.ValAddress))
// 	// After we set it, we should be able to find it.
// 	k.SetReport(ctx, 42, types.NewReport(bandtesting.Alice.ValAddress, true, nil))
// 	require.True(t, k.HasReport(ctx, 42, bandtesting.Alice.ValAddress))
// }

// func TestAddReportSuccess(t *testing.T) {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	k.SetRequest(ctx, 1, defaultRequest())
// 	err := k.AddReport(ctx, 1,
// 		bandtesting.Validators[0].ValAddress, true, []types.RawReport{
// 			types.NewRawReport(42, 0, []byte("data1/1")),
// 			types.NewRawReport(43, 1, []byte("data2/1")),
// 		},
// 	)
// 	require.NoError(t, err)
// 	require.Equal(t, []types.Report{
// 		types.NewReport(bandtesting.Validators[0].ValAddress, true, []types.RawReport{
// 			types.NewRawReport(42, 0, []byte("data1/1")),
// 			types.NewRawReport(43, 1, []byte("data2/1")),
// 		}),
// 	}, k.GetReports(ctx, 1))
// }

// func TestReportOnNonExistingRequest(t *testing.T) {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	err := k.AddReport(ctx, 1,
// 		bandtesting.Validators[0].ValAddress, true, []types.RawReport{
// 			types.NewRawReport(42, 0, []byte("data1/1")),
// 			types.NewRawReport(43, 1, []byte("data2/1")),
// 		},
// 	)
// 	require.ErrorIs(t, err, types.ErrRequestNotFound)
// }

// func TestReportByNotRequestedValidator(t *testing.T) {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	k.SetRequest(ctx, 1, defaultRequest())
// 	err := k.AddReport(ctx, 1,
// 		bandtesting.Alice.ValAddress, true, []types.RawReport{
// 			types.NewRawReport(42, 0, []byte("data1/1")),
// 			types.NewRawReport(43, 1, []byte("data2/1")),
// 		},
// 	)
// 	require.ErrorIs(t, err, types.ErrValidatorNotRequested)
// }

// func TestDuplicateReport(t *testing.T) {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	k.SetRequest(ctx, 1, defaultRequest())
// 	err := k.AddReport(ctx, 1,
// 		bandtesting.Validators[0].ValAddress, true, []types.RawReport{
// 			types.NewRawReport(42, 0, []byte("data1/1")),
// 			types.NewRawReport(43, 1, []byte("data2/1")),
// 		},
// 	)
// 	require.NoError(t, err)
// 	err = k.AddReport(ctx, 1,
// 		bandtesting.Validators[0].ValAddress, true, []types.RawReport{
// 			types.NewRawReport(42, 0, []byte("data1/1")),
// 			types.NewRawReport(43, 1, []byte("data2/1")),
// 		},
// 	)
// 	require.ErrorIs(t, err, types.ErrValidatorAlreadyReported)
// }

// func TestReportInvalidDataSourceCount(t *testing.T) {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	k.SetRequest(ctx, 1, defaultRequest())
// 	err := k.AddReport(ctx, 1,
// 		bandtesting.Validators[0].ValAddress, true, []types.RawReport{
// 			types.NewRawReport(42, 0, []byte("data1/1")),
// 		},
// 	)
// 	require.ErrorIs(t, err, types.ErrInvalidReportSize)
// }

// func TestReportInvalidExternalIDs(t *testing.T) {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	k.SetRequest(ctx, 1, defaultRequest())
// 	err := k.AddReport(ctx, 1,
// 		bandtesting.Validators[0].ValAddress, true, []types.RawReport{
// 			types.NewRawReport(42, 0, []byte("data1/1")),
// 			types.NewRawReport(44, 1, []byte("data2/1")), // BAD EXTERNAL ID!
// 		},
// 	)
// 	require.ErrorIs(t, err, types.ErrRawRequestNotFound)
// }

// func TestGetReportCount(t *testing.T) {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	// We start by setting some arbitrary reports.
// 	k.SetReport(ctx, types.RequestID(1), types.NewReport(bandtesting.Alice.ValAddress, true, []types.RawReport{}))
// 	k.SetReport(ctx, types.RequestID(1), types.NewReport(bandtesting.Bob.ValAddress, true, []types.RawReport{}))
// 	k.SetReport(ctx, types.RequestID(2), types.NewReport(bandtesting.Alice.ValAddress, true, []types.RawReport{}))
// 	k.SetReport(ctx, types.RequestID(2), types.NewReport(bandtesting.Bob.ValAddress, true, []types.RawReport{}))
// 	k.SetReport(ctx, types.RequestID(2), types.NewReport(bandtesting.Carol.ValAddress, true, []types.RawReport{}))
// 	// GetReportCount should return the correct values.
// 	require.Equal(t, uint64(2), k.GetReportCount(ctx, types.RequestID(1)))
// 	require.Equal(t, uint64(3), k.GetReportCount(ctx, types.RequestID(2)))
// }

// func TestDeleteReports(t *testing.T) {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	// We start by setting some arbitrary reports.
// 	k.SetReport(ctx, types.RequestID(1), types.NewReport(bandtesting.Alice.ValAddress, true, []types.RawReport{}))
// 	k.SetReport(ctx, types.RequestID(1), types.NewReport(bandtesting.Bob.ValAddress, true, []types.RawReport{}))
// 	k.SetReport(ctx, types.RequestID(2), types.NewReport(bandtesting.Alice.ValAddress, true, []types.RawReport{}))
// 	k.SetReport(ctx, types.RequestID(2), types.NewReport(bandtesting.Bob.ValAddress, true, []types.RawReport{}))
// 	k.SetReport(ctx, types.RequestID(2), types.NewReport(bandtesting.Carol.ValAddress, true, []types.RawReport{}))
// 	// All reports should exist on the state.
// 	require.True(t, k.HasReport(ctx, types.RequestID(1), bandtesting.Alice.ValAddress))
// 	require.True(t, k.HasReport(ctx, types.RequestID(1), bandtesting.Bob.ValAddress))
// 	require.True(t, k.HasReport(ctx, types.RequestID(2), bandtesting.Alice.ValAddress))
// 	require.True(t, k.HasReport(ctx, types.RequestID(2), bandtesting.Bob.ValAddress))
// 	require.True(t, k.HasReport(ctx, types.RequestID(2), bandtesting.Carol.ValAddress))
// 	// After we delete reports related to request#1, they must disappear.
// 	k.DeleteReports(ctx, types.RequestID(1))
// 	require.False(t, k.HasReport(ctx, types.RequestID(1), bandtesting.Alice.ValAddress))
// 	require.False(t, k.HasReport(ctx, types.RequestID(1), bandtesting.Bob.ValAddress))
// 	require.True(t, k.HasReport(ctx, types.RequestID(2), bandtesting.Alice.ValAddress))
// 	require.True(t, k.HasReport(ctx, types.RequestID(2), bandtesting.Bob.ValAddress))
// 	require.True(t, k.HasReport(ctx, types.RequestID(2), bandtesting.Carol.ValAddress))
// }
