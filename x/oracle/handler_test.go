package oracle_test

import (
	"bytes"
	gz "compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/bandprotocol/go-owasm/api"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	bandtesting "github.com/bandprotocol/chain/v2/testing"
	"github.com/bandprotocol/chain/v2/testing/testdata"
	"github.com/bandprotocol/chain/v2/x/oracle"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

func TestCreateDataSourceSuccess(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, false)
	k := app.OracleKeeper

	dsCount := k.GetDataSourceCount(ctx)
	treasury := bandtesting.Treasury.Address
	owner := bandtesting.Owner.Address
	name := "data_source_1"
	description := "description"
	executable := []byte("executable")
	executableHash := sha256.Sum256(executable)
	filename := hex.EncodeToString(executableHash[:])
	msg := types.NewMsgCreateDataSource(
		name,
		description,
		executable,
		bandtesting.EmptyCoins,
		treasury,
		owner,
		bandtesting.Alice.Address,
	)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	ds, err := k.GetDataSource(ctx, types.DataSourceID(dsCount+1))
	require.NoError(t, err)
	require.Equal(
		t,
		types.NewDataSource(bandtesting.Owner.Address, name, description, filename, bandtesting.EmptyCoins, treasury),
		ds,
	)
	event := abci.Event{
		Type: types.EventTypeCreateDataSource,
		Attributes: []abci.EventAttribute{
			{Key: types.AttributeKeyID, Value: fmt.Sprintf("%d", dsCount+1)},
		},
	}
	require.Equal(t, event, res.Events[0])
}

func TestCreateGzippedExecutableDataSourceFail(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, true)
	k := app.OracleKeeper

	treasury := bandtesting.Treasury.Address
	owner := bandtesting.Owner.Address
	name := "data_source_1"
	description := "description"
	executable := []byte("executable")
	var buf bytes.Buffer
	zw := gz.NewWriter(&buf)
	_, err := zw.Write(executable)
	require.NoError(t, err)
	zw.Close()
	sender := bandtesting.Alice.Address
	msg := types.NewMsgCreateDataSource(
		name,
		description,
		buf.Bytes()[:5],
		bandtesting.EmptyCoins,
		treasury,
		owner,
		sender,
	)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.ErrorIs(t, err, types.ErrUncompressionFailed)
	require.Nil(t, res)
}

func TestEditDataSourceSuccess(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, false)
	k := app.OracleKeeper

	newName := "beeb"
	newDescription := "new_description"
	newExecutable := []byte("executable2")
	newExecutableHash := sha256.Sum256(newExecutable)
	newFilename := hex.EncodeToString(newExecutableHash[:])
	msg := types.NewMsgEditDataSource(
		1,
		newName,
		newDescription,
		newExecutable,
		bandtesting.Coins1000000uband,
		bandtesting.Treasury.Address,
		bandtesting.Alice.Address,
		bandtesting.Owner.Address,
	)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	ds, err := k.GetDataSource(ctx, 1)
	require.NoError(t, err)
	require.Equal(
		t,
		types.NewDataSource(
			bandtesting.Alice.Address,
			newName,
			newDescription,
			newFilename,
			bandtesting.Coins1000000uband,
			bandtesting.Treasury.Address,
		),
		ds,
	)
	event := abci.Event{
		Type:       types.EventTypeEditDataSource,
		Attributes: []abci.EventAttribute{{Key: types.AttributeKeyID, Value: "1"}},
	}
	require.Equal(t, event, res.Events[0])
}

func TestEditDataSourceFail(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, false)
	k := app.OracleKeeper

	newName := "beeb"
	newDescription := "new_description"
	newExecutable := []byte("executable2")
	// Bad ID
	msg := types.NewMsgEditDataSource(
		42,
		newName,
		newDescription,
		newExecutable,
		bandtesting.EmptyCoins,
		bandtesting.Treasury.Address,
		bandtesting.Owner.Address,
		bandtesting.Owner.Address,
	)
	res, err := oracle.NewHandler(k)(ctx, msg)
	bandtesting.CheckErrorf(t, err, types.ErrDataSourceNotFound, "id: 42")
	require.Nil(t, res)
	// Not owner
	msg = types.NewMsgEditDataSource(
		1,
		newName,
		newDescription,
		newExecutable,
		bandtesting.EmptyCoins,
		bandtesting.Treasury.Address,
		bandtesting.Owner.Address,
		bandtesting.Bob.Address,
	)
	res, err = oracle.NewHandler(k)(ctx, msg)
	require.ErrorIs(t, err, types.ErrEditorNotAuthorized)
	require.Nil(t, res)
	// Bad Gzip
	var buf bytes.Buffer
	zw := gz.NewWriter(&buf)
	_, err = zw.Write(newExecutable)
	require.NoError(t, err)
	zw.Close()
	msg = types.NewMsgEditDataSource(
		1,
		newName,
		newDescription,
		buf.Bytes()[:5],
		bandtesting.EmptyCoins,
		bandtesting.Treasury.Address,
		bandtesting.Owner.Address,
		bandtesting.Owner.Address,
	)
	res, err = oracle.NewHandler(k)(ctx, msg)
	require.ErrorIs(t, err, types.ErrUncompressionFailed)
	require.Nil(t, res)
}

func TestCreateOracleScriptSuccess(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, false)
	k := app.OracleKeeper

	osCount := k.GetOracleScriptCount(ctx)
	name := "os_1"
	description := "beeb"
	code := testdata.WasmExtra1
	schema := "schema"
	url := "url"
	msg := types.NewMsgCreateOracleScript(
		name,
		description,
		schema,
		url,
		code,
		bandtesting.Owner.Address,
		bandtesting.Alice.Address,
	)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	os, err := k.GetOracleScript(ctx, types.OracleScriptID(osCount+1))
	require.NoError(t, err)
	require.Equal(
		t,
		types.NewOracleScript(
			bandtesting.Owner.Address,
			name,
			description,
			testdata.WasmExtra1FileName,
			schema,
			url,
		),
		os,
	)

	event := abci.Event{
		Type: types.EventTypeCreateOracleScript,
		Attributes: []abci.EventAttribute{
			{Key: types.AttributeKeyID, Value: fmt.Sprintf("%d", osCount+1)},
		},
	}
	require.Equal(t, event, res.Events[0])
}

func TestCreateGzippedOracleScriptSuccess(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, false)
	k := app.OracleKeeper

	osCount := k.GetOracleScriptCount(ctx)
	name := "os_1"
	description := "beeb"
	schema := "schema"
	url := "url"
	var buf bytes.Buffer
	zw := gz.NewWriter(&buf)
	_, err := zw.Write(testdata.WasmExtra1)
	require.NoError(t, err)
	zw.Close()
	msg := types.NewMsgCreateOracleScript(
		name,
		description,
		schema,
		url,
		buf.Bytes(),
		bandtesting.Owner.Address,
		bandtesting.Alice.Address,
	)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	os, err := k.GetOracleScript(ctx, types.OracleScriptID(osCount+1))
	require.NoError(t, err)
	require.Equal(
		t,
		types.NewOracleScript(
			bandtesting.Owner.Address,
			name,
			description,
			testdata.WasmExtra1FileName,
			schema,
			url,
		),
		os,
	)

	event := abci.Event{
		Type: types.EventTypeCreateOracleScript,
		Attributes: []abci.EventAttribute{
			{Key: types.AttributeKeyID, Value: fmt.Sprintf("%d", osCount+1)},
		},
	}
	require.Equal(t, event, res.Events[0])
}

func TestCreateOracleScriptFail(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, false)
	k := app.OracleKeeper

	name := "os_1"
	description := "beeb"
	schema := "schema"
	url := "url"
	// Bad Owasm code
	msg := types.NewMsgCreateOracleScript(
		name,
		description,
		schema,
		url,
		[]byte("BAD"),
		bandtesting.Owner.Address,
		bandtesting.Alice.Address,
	)
	res, err := oracle.NewHandler(k)(ctx, msg)
	bandtesting.CheckErrorf(t, err, types.ErrOwasmCompilation, "caused by %s", api.ErrValidation)
	require.Nil(t, res)
	// Bad Gzip
	var buf bytes.Buffer
	zw := gz.NewWriter(&buf)
	_, err = zw.Write(testdata.WasmExtra1)
	require.NoError(t, err)
	zw.Close()
	msg = types.NewMsgCreateOracleScript(
		name,
		description,
		schema,
		url,
		buf.Bytes()[:5],
		bandtesting.Owner.Address,
		bandtesting.Alice.Address,
	)
	res, err = oracle.NewHandler(k)(ctx, msg)
	require.ErrorIs(t, err, types.ErrUncompressionFailed)
	require.Nil(t, res)
}

func TestEditOracleScriptSuccess(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, false)
	k := app.OracleKeeper

	newName := "os_2"
	newDescription := "beebbeeb"
	newCode := testdata.WasmExtra2
	newSchema := "new_schema"
	newURL := "new_url"
	msg := types.NewMsgEditOracleScript(
		1,
		newName,
		newDescription,
		newSchema,
		newURL,
		newCode,
		bandtesting.Alice.Address,
		bandtesting.Owner.Address,
	)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	os, err := k.GetOracleScript(ctx, 1)
	require.NoError(t, err)
	require.Equal(
		t,
		types.NewOracleScript(
			bandtesting.Alice.Address,
			newName,
			newDescription,
			testdata.WasmExtra2FileName,
			newSchema,
			newURL,
		),
		os,
	)

	event := abci.Event{
		Type:       types.EventTypeEditOracleScript,
		Attributes: []abci.EventAttribute{{Key: types.AttributeKeyID, Value: "1"}},
	}
	require.Equal(t, event, res.Events[0])
}

func TestEditOracleScriptFail(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, false)
	k := app.OracleKeeper

	newName := "os_2"
	newDescription := "beebbeeb"
	newCode := testdata.WasmExtra2
	newSchema := "new_schema"
	newURL := "new_url"
	// Bad ID
	msg := types.NewMsgEditOracleScript(
		999,
		newName,
		newDescription,
		newSchema,
		newURL,
		newCode,
		bandtesting.Owner.Address,
		bandtesting.Owner.Address,
	)
	res, err := oracle.NewHandler(k)(ctx, msg)
	bandtesting.CheckErrorf(t, err, types.ErrOracleScriptNotFound, "id: 999")
	require.Nil(t, res)
	// Not owner
	msg = types.NewMsgEditOracleScript(
		1,
		newName,
		newDescription,
		newSchema,
		newURL,
		newCode,
		bandtesting.Owner.Address,
		bandtesting.Bob.Address,
	)
	res, err = oracle.NewHandler(k)(ctx, msg)
	require.EqualError(t, err, "editor not authorized")
	require.Nil(t, res)
	// Bad Owasm code
	msg = types.NewMsgEditOracleScript(
		1,
		newName,
		newDescription,
		newSchema,
		newURL,
		[]byte("BAD_CODE"),
		bandtesting.Owner.Address,
		bandtesting.Owner.Address,
	)
	res, err = oracle.NewHandler(k)(ctx, msg)
	bandtesting.CheckErrorf(t, err, types.ErrOwasmCompilation, "caused by %s", api.ErrValidation)
	require.Nil(t, res)
	// Bad Gzip
	var buf bytes.Buffer
	zw := gz.NewWriter(&buf)
	_, err = zw.Write(testdata.WasmExtra2)
	require.NoError(t, err)
	zw.Close()
	msg = types.NewMsgEditOracleScript(
		1,
		newName,
		newDescription,
		newSchema,
		newURL,
		buf.Bytes()[:5],
		bandtesting.Owner.Address,
		bandtesting.Owner.Address,
	)
	res, err = oracle.NewHandler(k)(ctx, msg)
	require.ErrorIs(t, err, types.ErrUncompressionFailed)
	require.Nil(t, res)
}

func TestRequestDataSuccess(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, true)
	k := app.OracleKeeper

	ctx = ctx.WithBlockHeight(124).WithBlockTime(bandtesting.ParseTime(1581589790))
	msg := types.NewMsgRequestData(
		1,
		[]byte("beeb"),
		2,
		2,
		"CID",
		bandtesting.Coins100000000uband,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.FeePayer.Address,
	)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, types.NewRequest(
		1,
		[]byte("beeb"),
		[]sdk.ValAddress{bandtesting.Validators[2].ValAddress, bandtesting.Validators[0].ValAddress},
		2,
		124,
		bandtesting.ParseTime(1581589790),
		"CID",
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("beeb")),
			types.NewRawRequest(2, 2, []byte("beeb")),
			types.NewRawRequest(3, 3, []byte("beeb")),
		},
		nil,
		bandtesting.TestDefaultExecuteGas,
	), k.MustGetRequest(ctx, 1))
	event := abci.Event{
		Type: authtypes.EventTypeCoinSpent,
		Attributes: []abci.EventAttribute{
			{Key: authtypes.AttributeKeySpender, Value: bandtesting.FeePayer.Address.String()},
			{Key: sdk.AttributeKeyAmount, Value: "2000000uband"},
		},
	}
	require.Equal(t, event, res.Events[0])
	require.Equal(t, event, res.Events[4])
	require.Equal(t, event, res.Events[8])
	event = abci.Event{
		Type: authtypes.EventTypeCoinReceived,
		Attributes: []abci.EventAttribute{
			{Key: authtypes.AttributeKeyReceiver, Value: bandtesting.Treasury.Address.String()},
			{Key: sdk.AttributeKeyAmount, Value: "2000000uband"},
		},
	}
	require.Equal(t, event, res.Events[1])
	require.Equal(t, event, res.Events[5])
	require.Equal(t, event, res.Events[9])
	event = abci.Event{
		Type: authtypes.EventTypeTransfer,
		Attributes: []abci.EventAttribute{
			{Key: authtypes.AttributeKeyRecipient, Value: bandtesting.Treasury.Address.String()},
			{Key: authtypes.AttributeKeySender, Value: bandtesting.FeePayer.Address.String()},
			{Key: sdk.AttributeKeyAmount, Value: "2000000uband"},
		},
	}
	require.Equal(t, event, res.Events[2])
	require.Equal(t, event, res.Events[6])
	require.Equal(t, event, res.Events[10])
	event = abci.Event{
		Type: sdk.EventTypeMessage,
		Attributes: []abci.EventAttribute{
			{Key: authtypes.AttributeKeySender, Value: bandtesting.FeePayer.Address.String()},
		},
	}
	require.Equal(t, event, res.Events[3])
	require.Equal(t, event, res.Events[7])
	require.Equal(t, event, res.Events[11])

	event = abci.Event{
		Type: types.EventTypeRequest,
		Attributes: []abci.EventAttribute{
			{Key: types.AttributeKeyID, Value: "1"},
			{Key: types.AttributeKeyClientID, Value: "CID"},
			{Key: types.AttributeKeyOracleScriptID, Value: "1"},
			{Key: types.AttributeKeyCalldata, Value: "62656562"}, // "beeb" in hex
			{Key: types.AttributeKeyAskCount, Value: "2"},
			{Key: types.AttributeKeyMinCount, Value: "2"},
			{Key: types.AttributeKeyGasUsed, Value: "5294700000"},
			{Key: types.AttributeKeyTotalFees, Value: "6000000uband"},
			{Key: types.AttributeKeyValidator, Value: bandtesting.Validators[2].ValAddress.String()},
			{Key: types.AttributeKeyValidator, Value: bandtesting.Validators[0].ValAddress.String()},
		},
	}
	require.Equal(t, event, res.Events[12])
	event = abci.Event{
		Type: types.EventTypeRawRequest,
		Attributes: []abci.EventAttribute{
			{Key: types.AttributeKeyDataSourceID, Value: "1"},
			{Key: types.AttributeKeyDataSourceHash, Value: bandtesting.DataSources[1].Filename},
			{Key: types.AttributeKeyExternalID, Value: "1"},
			{Key: types.AttributeKeyCalldata, Value: "beeb"},
			{Key: types.AttributeKeyFee, Value: "1000000uband"},
		},
	}
	require.Equal(t, event, res.Events[13])
	event = abci.Event{
		Type: types.EventTypeRawRequest,
		Attributes: []abci.EventAttribute{
			{Key: types.AttributeKeyDataSourceID, Value: "2"},
			{Key: types.AttributeKeyDataSourceHash, Value: bandtesting.DataSources[2].Filename},
			{Key: types.AttributeKeyExternalID, Value: "2"},
			{Key: types.AttributeKeyCalldata, Value: "beeb"},
			{Key: types.AttributeKeyFee, Value: "1000000uband"},
		},
	}
	require.Equal(t, event, res.Events[14])
	event = abci.Event{
		Type: types.EventTypeRawRequest,
		Attributes: []abci.EventAttribute{
			{Key: types.AttributeKeyDataSourceID, Value: "3"},
			{Key: types.AttributeKeyDataSourceHash, Value: bandtesting.DataSources[3].Filename},
			{Key: types.AttributeKeyExternalID, Value: "3"},
			{Key: types.AttributeKeyCalldata, Value: "beeb"},
			{Key: types.AttributeKeyFee, Value: "1000000uband"},
		},
	}
	require.Equal(t, event, res.Events[15])
}

func TestRequestDataFail(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, false)
	k := app.OracleKeeper

	// No active oracle validators
	res, err := oracle.NewHandler(
		k,
	)(
		ctx,
		types.NewMsgRequestData(
			1,
			[]byte("beeb"),
			2,
			2,
			"CID",
			bandtesting.Coins100000000uband,
			bandtesting.TestDefaultPrepareGas,
			bandtesting.TestDefaultExecuteGas,
			bandtesting.FeePayer.Address,
		),
	)
	bandtesting.CheckErrorf(t, err, types.ErrInsufficientValidators, "0 < 2")
	require.Nil(t, res)
	err = k.Activate(ctx, bandtesting.Validators[0].ValAddress)
	require.NoError(t, err)
	err = k.Activate(ctx, bandtesting.Validators[1].ValAddress)
	require.NoError(t, err)
	// Too large calldata
	res, err = oracle.NewHandler(
		k,
	)(
		ctx,
		types.NewMsgRequestData(
			1,
			[]byte(strings.Repeat("beeb", 2000)),
			2,
			2,
			"CID",
			bandtesting.Coins100000000uband,
			bandtesting.TestDefaultPrepareGas,
			bandtesting.TestDefaultExecuteGas,
			bandtesting.FeePayer.Address,
		),
	)
	bandtesting.CheckErrorf(t, err, types.ErrTooLargeCalldata, "got: 8000, max: 512")
	require.Nil(t, res)
	// Too high ask count
	res, err = oracle.NewHandler(
		k,
	)(
		ctx,
		types.NewMsgRequestData(
			1,
			[]byte("beeb"),
			3,
			2,
			"CID",
			bandtesting.Coins100000000uband,
			bandtesting.TestDefaultPrepareGas,
			bandtesting.TestDefaultExecuteGas,
			bandtesting.FeePayer.Address,
		),
	)
	bandtesting.CheckErrorf(t, err, types.ErrInsufficientValidators, "2 < 3")
	require.Nil(t, res)
	// Bad oracle script ID
	res, err = oracle.NewHandler(
		k,
	)(
		ctx,
		types.NewMsgRequestData(
			999,
			[]byte("beeb"),
			2,
			2,
			"CID",
			bandtesting.Coins100000000uband,
			bandtesting.TestDefaultPrepareGas,
			bandtesting.TestDefaultExecuteGas,
			bandtesting.FeePayer.Address,
		),
	)
	bandtesting.CheckErrorf(t, err, types.ErrOracleScriptNotFound, "id: 999")
	require.Nil(t, res)
	// Pay not enough fee
	res, err = oracle.NewHandler(
		k,
	)(
		ctx,
		types.NewMsgRequestData(
			1,
			[]byte("beeb"),
			2,
			2,
			"CID",
			bandtesting.EmptyCoins,
			bandtesting.TestDefaultPrepareGas,
			bandtesting.TestDefaultExecuteGas,
			bandtesting.FeePayer.Address,
		),
	)
	bandtesting.CheckErrorf(t, err, types.ErrNotEnoughFee, "require: 2000000uband, max: 0uband")
	require.Nil(t, res)
}

func TestReportSuccess(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, true)
	k := app.OracleKeeper

	// Set up a mock request asking 3 validators with min count 2.
	k.SetRequest(ctx, 42, types.NewRequest(
		1,
		[]byte("beeb"),
		[]sdk.ValAddress{
			bandtesting.Validators[2].ValAddress,
			bandtesting.Validators[1].ValAddress,
			bandtesting.Validators[0].ValAddress,
		},
		2,
		124,
		bandtesting.ParseTime(1581589790),
		"CID",
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("beeb")),
			types.NewRawRequest(2, 2, []byte("beeb")),
		},
		nil,
		0,
	))
	// Common raw reports for everyone.
	reports := []types.RawReport{types.NewRawReport(1, 0, []byte("data1")), types.NewRawReport(2, 0, []byte("data2"))}
	// Validators[0] reports data.
	res, err := oracle.NewHandler(k)(ctx, types.NewMsgReportData(42, reports, bandtesting.Validators[0].ValAddress))
	require.NoError(t, err)
	require.Equal(t, []types.RequestID{}, k.GetPendingResolveList(ctx))
	event := abci.Event{
		Type: types.EventTypeReport,
		Attributes: []abci.EventAttribute{
			{Key: types.AttributeKeyID, Value: "42"},
			{Key: types.AttributeKeyValidator, Value: bandtesting.Validators[0].ValAddress.String()},
		},
	}
	require.Equal(t, event, res.Events[0])
	// Validators[1] reports data. Now the request should move to pending resolve.
	res, err = oracle.NewHandler(k)(ctx, types.NewMsgReportData(42, reports, bandtesting.Validators[1].ValAddress))
	require.NoError(t, err)
	require.Equal(t, []types.RequestID{42}, k.GetPendingResolveList(ctx))
	event = abci.Event{
		Type: types.EventTypeReport,
		Attributes: []abci.EventAttribute{
			{Key: types.AttributeKeyID, Value: "42"},
			{Key: types.AttributeKeyValidator, Value: bandtesting.Validators[1].ValAddress.String()},
		},
	}
	require.Equal(t, event, res.Events[0])
	// Even if we resolve the request, Validators[2] should still be able to report.
	k.SetPendingResolveList(ctx, []types.RequestID{})
	k.ResolveSuccess(ctx, 42, []byte("RESOLVE_RESULT!"), 1234)
	res, err = oracle.NewHandler(k)(ctx, types.NewMsgReportData(42, reports, bandtesting.Validators[2].ValAddress))
	require.NoError(t, err)
	event = abci.Event{
		Type: types.EventTypeReport,
		Attributes: []abci.EventAttribute{
			{Key: types.AttributeKeyID, Value: "42"},
			{Key: types.AttributeKeyValidator, Value: bandtesting.Validators[2].ValAddress.String()},
		},
	}
	require.Equal(t, event, res.Events[0])
	// Check the reports of this request. We should see 3 reports, with report from Validators[2] comes after resolve.
	finalReport := k.GetReports(ctx, 42)
	require.Contains(t, finalReport, types.NewReport(bandtesting.Validators[0].ValAddress, true, reports))
	require.Contains(t, finalReport, types.NewReport(bandtesting.Validators[1].ValAddress, true, reports))
	require.Contains(t, finalReport, types.NewReport(bandtesting.Validators[2].ValAddress, false, reports))
}

func TestReportFail(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, true)
	k := app.OracleKeeper

	// Set up a mock request asking 3 validators with min count 2.
	k.SetRequest(ctx, 42, types.NewRequest(
		1,
		[]byte("beeb"),
		[]sdk.ValAddress{
			bandtesting.Validators[2].ValAddress,
			bandtesting.Validators[1].ValAddress,
			bandtesting.Validators[0].ValAddress,
		},
		2,
		124,
		bandtesting.ParseTime(1581589790),
		"CID",
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("beeb")),
			types.NewRawRequest(2, 2, []byte("beeb")),
		},
		nil,
		0,
	))
	// Common raw reports for everyone.
	reports := []types.RawReport{types.NewRawReport(1, 0, []byte("data1")), types.NewRawReport(2, 0, []byte("data2"))}
	// Bad ID
	res, err := oracle.NewHandler(k)(ctx, types.NewMsgReportData(999, reports, bandtesting.Validators[0].ValAddress))
	bandtesting.CheckErrorf(t, err, types.ErrRequestNotFound, "id: 999")
	require.Nil(t, res)
	// Not-asked validator
	res, err = oracle.NewHandler(k)(ctx, types.NewMsgReportData(42, reports, bandtesting.Alice.ValAddress))
	bandtesting.CheckErrorf(
		t,
		err,
		types.ErrValidatorNotRequested,
		"reqID: 42, val: %s",
		bandtesting.Alice.ValAddress.String(),
	)
	require.Nil(t, res)
	// Too large report data size
	res, err = oracle.NewHandler(
		k,
	)(
		ctx,
		types.NewMsgReportData(
			42,
			[]types.RawReport{
				types.NewRawReport(1, 0, []byte("data1")),
				types.NewRawReport(2, 0, []byte(strings.Repeat("data2", 2000))),
			},
			bandtesting.Validators[0].ValAddress,
		),
	)
	bandtesting.CheckErrorf(t, err, types.ErrTooLargeRawReportData, "got: 10000, max: 512")
	require.Nil(t, res)
	// Not having all raw reports
	res, err = oracle.NewHandler(
		k,
	)(
		ctx,
		types.NewMsgReportData(
			42,
			[]types.RawReport{types.NewRawReport(1, 0, []byte("data1"))},
			bandtesting.Validators[0].ValAddress,
		),
	)
	require.ErrorIs(t, err, types.ErrInvalidReportSize)
	require.Nil(t, res)
	// Incorrect external IDs
	res, err = oracle.NewHandler(
		k,
	)(
		ctx,
		types.NewMsgReportData(
			42,
			[]types.RawReport{types.NewRawReport(1, 0, []byte("data1")), types.NewRawReport(42, 0, []byte("data2"))},
			bandtesting.Validators[0].ValAddress,
		),
	)
	bandtesting.CheckErrorf(t, err, types.ErrRawRequestNotFound, "reqID: 42, extID: 42")
	require.Nil(t, res)
	// Request already expired
	k.SetRequestLastExpired(ctx, 42)
	res, err = oracle.NewHandler(k)(ctx, types.NewMsgReportData(42, reports, bandtesting.Validators[0].ValAddress))
	require.ErrorIs(t, err, types.ErrRequestAlreadyExpired)
	require.Nil(t, res)
}

func TestActivateSuccess(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, false)
	k := app.OracleKeeper

	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1000000))
	require.Equal(t,
		types.NewValidatorStatus(false, time.Time{}),
		k.GetValidatorStatus(ctx, bandtesting.Validators[0].ValAddress),
	)
	msg := types.NewMsgActivate(bandtesting.Validators[0].ValAddress)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	require.Equal(t,
		types.NewValidatorStatus(true, bandtesting.ParseTime(1000000)),
		k.GetValidatorStatus(ctx, bandtesting.Validators[0].ValAddress),
	)
	event := abci.Event{
		Type: types.EventTypeActivate,
		Attributes: []abci.EventAttribute{
			{Key: types.AttributeKeyValidator, Value: bandtesting.Validators[0].ValAddress.String()},
		},
	}
	require.Equal(t, event, res.Events[0])
}

func TestActivateFail(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, true)
	k := app.OracleKeeper

	msg := types.NewMsgActivate(bandtesting.Validators[0].ValAddress)
	// Already active.
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.ErrorIs(t, err, types.ErrValidatorAlreadyActive)
	require.Nil(t, res)
	// Too soon to activate.
	ctx = ctx.WithBlockTime(bandtesting.ParseTime(100000))
	k.MissReport(ctx, bandtesting.Validators[0].ValAddress, bandtesting.ParseTime(99999))
	ctx = ctx.WithBlockTime(bandtesting.ParseTime(100001))
	res, err = oracle.NewHandler(k)(ctx, msg)
	require.ErrorIs(t, err, types.ErrTooSoonToActivate)
	require.Nil(t, res)
	// OK
	ctx = ctx.WithBlockTime(bandtesting.ParseTime(200000))
	_, err = oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
}

func TestUpdateParamsSuccess(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, true)
	k := app.OracleKeeper

	expectedParams := types.Params{
		MaxRawRequestCount:      1,
		MaxAskCount:             10,
		MaxCalldataSize:         256,
		MaxReportDataSize:       512,
		ExpirationBlockCount:    30,
		BaseOwasmGas:            50000,
		PerValidatorRequestGas:  3000,
		SamplingTryCount:        3,
		OracleRewardPercentage:  50,
		InactivePenaltyDuration: 1000,
		IBCRequestEnabled:       true,
	}
	msg := types.NewMsgUpdateParams(k.GetAuthority(), expectedParams)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, expectedParams, k.GetParams(ctx))
	event := abci.Event{
		Type: types.EventTypeUpdateParams,
		Attributes: []abci.EventAttribute{
			{Key: types.AttributeKeyParams, Value: expectedParams.String()},
		},
	}
	require.Equal(t, event, res.Events[0])

	expectedParams = types.Params{
		MaxRawRequestCount:      2,
		MaxAskCount:             20,
		MaxCalldataSize:         512,
		MaxReportDataSize:       256,
		ExpirationBlockCount:    40,
		BaseOwasmGas:            0,
		PerValidatorRequestGas:  0,
		SamplingTryCount:        5,
		OracleRewardPercentage:  0,
		InactivePenaltyDuration: 0,
		IBCRequestEnabled:       false,
	}
	msg = types.NewMsgUpdateParams(k.GetAuthority(), expectedParams)
	res, err = oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, expectedParams, k.GetParams(ctx))
	event = abci.Event{
		Type: types.EventTypeUpdateParams,
		Attributes: []abci.EventAttribute{
			{Key: types.AttributeKeyParams, Value: expectedParams.String()},
		},
	}
	require.Equal(t, event, res.Events[0])
}

func TestUpdateParamsFail(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, true)
	k := app.OracleKeeper

	expectedParams := types.Params{
		MaxRawRequestCount:      1,
		MaxAskCount:             10,
		MaxCalldataSize:         256,
		MaxReportDataSize:       512,
		ExpirationBlockCount:    30,
		BaseOwasmGas:            50000,
		PerValidatorRequestGas:  3000,
		SamplingTryCount:        3,
		OracleRewardPercentage:  50,
		InactivePenaltyDuration: 1000,
		IBCRequestEnabled:       true,
	}
	msg := types.NewMsgUpdateParams("foo", expectedParams)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.ErrorContains(t, err, "invalid authority")
	require.Nil(t, res)

	expectedParams = types.Params{
		MaxRawRequestCount:      0,
		MaxAskCount:             10,
		MaxCalldataSize:         256,
		MaxReportDataSize:       512,
		ExpirationBlockCount:    30,
		BaseOwasmGas:            50000,
		PerValidatorRequestGas:  3000,
		SamplingTryCount:        3,
		OracleRewardPercentage:  50,
		InactivePenaltyDuration: 1000,
		IBCRequestEnabled:       true,
	}
	msg = types.NewMsgUpdateParams(k.GetAuthority(), expectedParams)
	res, err = oracle.NewHandler(k)(ctx, msg)
	require.ErrorContains(t, err, "max raw request count must be positive")
	require.Nil(t, res)
}
