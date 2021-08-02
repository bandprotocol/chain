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

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/bandprotocol/go-owasm/api"

	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/oracle"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

func TestCreateDataSourceSuccess(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	dsCount := k.GetDataSourceCount(ctx)
	treasury := testapp.Treasury.Address
	owner := testapp.Owner.Address
	name := "data_source_1"
	description := "description"
	executable := []byte("executable")
	executableHash := sha256.Sum256(executable)
	filename := hex.EncodeToString(executableHash[:])
	msg := types.NewMsgCreateDataSource(name, description, executable, testapp.EmptyCoins, treasury, owner, testapp.Alice.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	ds, err := k.GetDataSource(ctx, types.DataSourceID(dsCount+1))
	require.NoError(t, err)
	require.Equal(t, types.NewDataSource(testapp.Owner.Address, name, description, filename, testapp.EmptyCoins, treasury), ds)
	event := abci.Event{
		Type:       types.EventTypeCreateDataSource,
		Attributes: []abci.EventAttribute{{Key: []byte(types.AttributeKeyID), Value: []byte(fmt.Sprintf("%d", dsCount+1))}},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
}

func TestCreateGzippedExecutableDataSourceFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	treasury := testapp.Treasury.Address
	owner := testapp.Owner.Address
	name := "data_source_1"
	description := "description"
	executable := []byte("executable")
	var buf bytes.Buffer
	zw := gz.NewWriter(&buf)
	zw.Write(executable)
	zw.Close()
	sender := testapp.Alice.Address
	msg := types.NewMsgCreateDataSource(name, description, buf.Bytes()[:5], testapp.EmptyCoins, treasury, owner, sender)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.ErrorIs(t, err, types.ErrUncompressionFailed)
	require.Nil(t, res)
}

func TestEditDataSourceSuccess(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	newName := "beeb"
	newDescription := "new_description"
	newExecutable := []byte("executable2")
	newExecutableHash := sha256.Sum256(newExecutable)
	newFilename := hex.EncodeToString(newExecutableHash[:])
	msg := types.NewMsgEditDataSource(1, newName, newDescription, newExecutable, testapp.Coins1000000uband, testapp.Treasury.Address, testapp.Alice.Address, testapp.Owner.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	ds, err := k.GetDataSource(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, types.NewDataSource(testapp.Alice.Address, newName, newDescription, newFilename, testapp.Coins1000000uband, testapp.Treasury.Address), ds)
	event := abci.Event{
		Type:       types.EventTypeEditDataSource,
		Attributes: []abci.EventAttribute{{Key: []byte(types.AttributeKeyID), Value: []byte("1")}},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
}

func TestEditDataSourceFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	newName := "beeb"
	newDescription := "new_description"
	newExecutable := []byte("executable2")
	// Bad ID
	msg := types.NewMsgEditDataSource(42, newName, newDescription, newExecutable, testapp.EmptyCoins, testapp.Treasury.Address, testapp.Owner.Address, testapp.Owner.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	testapp.CheckErrorf(t, err, types.ErrDataSourceNotFound, "id: 42")
	require.Nil(t, res)
	// Not owner
	msg = types.NewMsgEditDataSource(1, newName, newDescription, newExecutable, testapp.EmptyCoins, testapp.Treasury.Address, testapp.Owner.Address, testapp.Bob.Address)
	res, err = oracle.NewHandler(k)(ctx, msg)
	require.ErrorIs(t, err, types.ErrEditorNotAuthorized)
	require.Nil(t, res)
	// Bad Gzip
	var buf bytes.Buffer
	zw := gz.NewWriter(&buf)
	zw.Write(newExecutable)
	zw.Close()
	msg = types.NewMsgEditDataSource(1, newName, newDescription, buf.Bytes()[:5], testapp.EmptyCoins, testapp.Treasury.Address, testapp.Owner.Address, testapp.Owner.Address)
	res, err = oracle.NewHandler(k)(ctx, msg)
	require.ErrorIs(t, err, types.ErrUncompressionFailed)
	require.Nil(t, res)
}

func TestCreateOracleScriptSuccess(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	osCount := k.GetOracleScriptCount(ctx)
	name := "os_1"
	description := "beeb"
	code := testapp.WasmExtra1
	schema := "schema"
	url := "url"
	msg := types.NewMsgCreateOracleScript(name, description, schema, url, code, testapp.Owner.Address, testapp.Alice.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	os, err := k.GetOracleScript(ctx, types.OracleScriptID(osCount+1))
	require.NoError(t, err)
	require.Equal(t, types.NewOracleScript(testapp.Owner.Address, name, description, testapp.WasmExtra1FileName, schema, url), os)

	event := abci.Event{
		Type:       types.EventTypeCreateOracleScript,
		Attributes: []abci.EventAttribute{{Key: []byte(types.AttributeKeyID), Value: []byte(fmt.Sprintf("%d", osCount+1))}},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
}

func TestCreateGzippedOracleScriptSuccess(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	osCount := k.GetOracleScriptCount(ctx)
	name := "os_1"
	description := "beeb"
	schema := "schema"
	url := "url"
	var buf bytes.Buffer
	zw := gz.NewWriter(&buf)
	zw.Write(testapp.WasmExtra1)
	zw.Close()
	msg := types.NewMsgCreateOracleScript(name, description, schema, url, buf.Bytes(), testapp.Owner.Address, testapp.Alice.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	os, err := k.GetOracleScript(ctx, types.OracleScriptID(osCount+1))
	require.NoError(t, err)
	require.Equal(t, types.NewOracleScript(testapp.Owner.Address, name, description, testapp.WasmExtra1FileName, schema, url), os)

	event := abci.Event{
		Type:       types.EventTypeCreateOracleScript,
		Attributes: []abci.EventAttribute{{Key: []byte(types.AttributeKeyID), Value: []byte(fmt.Sprintf("%d", osCount+1))}},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
}

func TestCreateOracleScriptFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	name := "os_1"
	description := "beeb"
	schema := "schema"
	url := "url"
	// Bad Owasm code
	msg := types.NewMsgCreateOracleScript(name, description, schema, url, []byte("BAD"), testapp.Owner.Address, testapp.Alice.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	testapp.CheckErrorf(t, err, types.ErrOwasmCompilation, "caused by %s", api.ErrValidation)
	require.Nil(t, res)
	// Bad Gzip
	var buf bytes.Buffer
	zw := gz.NewWriter(&buf)
	zw.Write(testapp.WasmExtra1)
	zw.Close()
	msg = types.NewMsgCreateOracleScript(name, description, schema, url, buf.Bytes()[:5], testapp.Owner.Address, testapp.Alice.Address)
	res, err = oracle.NewHandler(k)(ctx, msg)
	require.ErrorIs(t, err, types.ErrUncompressionFailed)
	require.Nil(t, res)
}

func TestEditOracleScriptSuccess(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	newName := "os_2"
	newDescription := "beebbeeb"
	newCode := testapp.WasmExtra2
	newSchema := "new_schema"
	newURL := "new_url"
	msg := types.NewMsgEditOracleScript(1, newName, newDescription, newSchema, newURL, newCode, testapp.Alice.Address, testapp.Owner.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	os, err := k.GetOracleScript(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, types.NewOracleScript(testapp.Alice.Address, newName, newDescription, testapp.WasmExtra2FileName, newSchema, newURL), os)

	event := abci.Event{
		Type:       types.EventTypeEditOracleScript,
		Attributes: []abci.EventAttribute{{Key: []byte(types.AttributeKeyID), Value: []byte("1")}},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
}

func TestEditOracleScriptFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	newName := "os_2"
	newDescription := "beebbeeb"
	newCode := testapp.WasmExtra2
	newSchema := "new_schema"
	newURL := "new_url"
	// Bad ID
	msg := types.NewMsgEditOracleScript(999, newName, newDescription, newSchema, newURL, newCode, testapp.Owner.Address, testapp.Owner.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	testapp.CheckErrorf(t, err, types.ErrOracleScriptNotFound, "id: 999")
	require.Nil(t, res)
	// Not owner
	msg = types.NewMsgEditOracleScript(1, newName, newDescription, newSchema, newURL, newCode, testapp.Owner.Address, testapp.Bob.Address)
	res, err = oracle.NewHandler(k)(ctx, msg)
	require.EqualError(t, err, "editor not authorized")
	require.Nil(t, res)
	// Bad Owasm code
	msg = types.NewMsgEditOracleScript(1, newName, newDescription, newSchema, newURL, []byte("BAD_CODE"), testapp.Owner.Address, testapp.Owner.Address)
	res, err = oracle.NewHandler(k)(ctx, msg)
	testapp.CheckErrorf(t, err, types.ErrOwasmCompilation, "caused by %s", api.ErrValidation)
	require.Nil(t, res)
	// Bad Gzip
	var buf bytes.Buffer
	zw := gz.NewWriter(&buf)
	zw.Write(testapp.WasmExtra2)
	zw.Close()
	msg = types.NewMsgEditOracleScript(1, newName, newDescription, newSchema, newURL, buf.Bytes()[:5], testapp.Owner.Address, testapp.Owner.Address)
	res, err = oracle.NewHandler(k)(ctx, msg)
	require.ErrorIs(t, err, types.ErrUncompressionFailed)
	require.Nil(t, res)
}

func TestRequestDataSuccess(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockHeight(124).WithBlockTime(testapp.ParseTime(1581589790))
	msg := types.NewMsgRequestData(1, []byte("beeb"), 2, 2, "CID", testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.FeePayer.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, types.NewRequest(
		1,
		[]byte("beeb"),
		[]sdk.ValAddress{testapp.Validators[2].ValAddress, testapp.Validators[0].ValAddress},
		2,
		124,
		testapp.ParseTime(1581589790),
		"CID",
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("beeb")),
			types.NewRawRequest(2, 2, []byte("beeb")),
			types.NewRawRequest(3, 3, []byte("beeb")),
		},
		nil,
		uint64(testapp.TestDefaultExecuteGas),
	), k.MustGetRequest(ctx, 1))

	event := abci.Event{
		Type: authtypes.EventTypeTransfer,
		Attributes: []abci.EventAttribute{
			{Key: []byte(authtypes.AttributeKeyRecipient), Value: []byte(testapp.Treasury.Address.String())},
			{Key: []byte(authtypes.AttributeKeySender), Value: []byte(testapp.FeePayer.Address.String())},
			{Key: []byte(sdk.AttributeKeyAmount), Value: []byte("2000000uband")},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
	require.Equal(t, abci.Event(event), res.Events[2])
	require.Equal(t, abci.Event(event), res.Events[4])
	event = abci.Event{
		Type: sdk.EventTypeMessage,
		Attributes: []abci.EventAttribute{
			{Key: []byte(authtypes.AttributeKeySender), Value: []byte(testapp.FeePayer.Address.String())},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[1])
	require.Equal(t, abci.Event(event), res.Events[3])
	require.Equal(t, abci.Event(event), res.Events[5])

	event = abci.Event{
		Type: types.EventTypeRequest,
		Attributes: []abci.EventAttribute{
			{Key: []byte(types.AttributeKeyID), Value: []byte("1")},
			{Key: []byte(types.AttributeKeyClientID), Value: []byte("CID")},
			{Key: []byte(types.AttributeKeyOracleScriptID), Value: []byte("1")},
			{Key: []byte(types.AttributeKeyCalldata), Value: []byte("62656562")}, // "beeb" in hex
			{Key: []byte(types.AttributeKeyAskCount), Value: []byte("2")},
			{Key: []byte(types.AttributeKeyMinCount), Value: []byte("2")},
			{Key: []byte(types.AttributeKeyGasUsed), Value: []byte("785")},
			{Key: []byte(types.AttributeKeyTotalFees), Value: []byte("6000000uband")},
			{Key: []byte(types.AttributeKeyValidator), Value: []byte(testapp.Validators[2].ValAddress.String())},
			{Key: []byte(types.AttributeKeyValidator), Value: []byte(testapp.Validators[0].ValAddress.String())},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[6])
	event = abci.Event{
		Type: types.EventTypeRawRequest,
		Attributes: []abci.EventAttribute{
			{Key: []byte(types.AttributeKeyDataSourceID), Value: []byte("1")},
			{Key: []byte(types.AttributeKeyDataSourceHash), Value: []byte(testapp.DataSources[1].Filename)},
			{Key: []byte(types.AttributeKeyExternalID), Value: []byte("1")},
			{Key: []byte(types.AttributeKeyCalldata), Value: []byte("beeb")},
			{Key: []byte(types.AttributeKeyFee), Value: []byte("1000000uband")},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[7])
	event = abci.Event{
		Type: types.EventTypeRawRequest,
		Attributes: []abci.EventAttribute{
			{Key: []byte(types.AttributeKeyDataSourceID), Value: []byte("2")},
			{Key: []byte(types.AttributeKeyDataSourceHash), Value: []byte(testapp.DataSources[2].Filename)},
			{Key: []byte(types.AttributeKeyExternalID), Value: []byte("2")},
			{Key: []byte(types.AttributeKeyCalldata), Value: []byte("beeb")},
			{Key: []byte(types.AttributeKeyFee), Value: []byte("1000000uband")},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[8])
	event = abci.Event{
		Type: types.EventTypeRawRequest,
		Attributes: []abci.EventAttribute{
			{Key: []byte(types.AttributeKeyDataSourceID), Value: []byte("3")},
			{Key: []byte(types.AttributeKeyDataSourceHash), Value: []byte(testapp.DataSources[3].Filename)},
			{Key: []byte(types.AttributeKeyExternalID), Value: []byte("3")},
			{Key: []byte(types.AttributeKeyCalldata), Value: []byte("beeb")},
			{Key: []byte(types.AttributeKeyFee), Value: []byte("1000000uband")},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[9])
}

func TestRequestDataFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	// No active oracle validators
	res, err := oracle.NewHandler(k)(ctx, types.NewMsgRequestData(1, []byte("beeb"), 2, 2, "CID", testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.FeePayer.Address))
	testapp.CheckErrorf(t, err, types.ErrInsufficientValidators, "0 < 2")
	require.Nil(t, res)
	k.Activate(ctx, testapp.Validators[0].ValAddress)
	k.Activate(ctx, testapp.Validators[1].ValAddress)
	// Too large calldata
	res, err = oracle.NewHandler(k)(ctx, types.NewMsgRequestData(1, []byte(strings.Repeat("beeb", 2000)), 2, 2, "CID", testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.FeePayer.Address))
	testapp.CheckErrorf(t, err, types.ErrTooLargeCalldata, "got: 8000, max: 256")
	require.Nil(t, res)
	// Too high ask count
	res, err = oracle.NewHandler(k)(ctx, types.NewMsgRequestData(1, []byte("beeb"), 3, 2, "CID", testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.FeePayer.Address))
	testapp.CheckErrorf(t, err, types.ErrInsufficientValidators, "2 < 3")
	require.Nil(t, res)
	// Bad oracle script ID
	res, err = oracle.NewHandler(k)(ctx, types.NewMsgRequestData(999, []byte("beeb"), 2, 2, "CID", testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.FeePayer.Address))
	testapp.CheckErrorf(t, err, types.ErrOracleScriptNotFound, "id: 999")
	require.Nil(t, res)
	// Pay not enough fee
	res, err = oracle.NewHandler(k)(ctx, types.NewMsgRequestData(1, []byte("beeb"), 2, 2, "CID", testapp.EmptyCoins, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.FeePayer.Address))
	testapp.CheckErrorf(t, err, types.ErrNotEnoughFee, "require: 2000000uband, max: 0uband")
	require.Nil(t, res)
}

func TestReportSuccess(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	// Set up a mock request asking 3 validators with min count 2.
	k.SetRequest(ctx, 42, types.NewRequest(
		1,
		[]byte("beeb"),
		[]sdk.ValAddress{testapp.Validators[2].ValAddress, testapp.Validators[1].ValAddress, testapp.Validators[0].ValAddress},
		2,
		124,
		testapp.ParseTime(1581589790),
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
	res, err := oracle.NewHandler(k)(ctx, types.NewMsgReportData(42, reports, testapp.Validators[0].ValAddress, testapp.Validators[0].Address))
	require.NoError(t, err)
	require.Equal(t, []types.RequestID{}, k.GetPendingResolveList(ctx))
	event := abci.Event{
		Type: types.EventTypeReport,
		Attributes: []abci.EventAttribute{
			{Key: []byte(types.AttributeKeyID), Value: []byte("42")},
			{Key: []byte(types.AttributeKeyValidator), Value: []byte(testapp.Validators[0].ValAddress.String())},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
	// Validators[1] reports data. Now the request should move to pending resolve.
	res, err = oracle.NewHandler(k)(ctx, types.NewMsgReportData(42, reports, testapp.Validators[1].ValAddress, testapp.Validators[1].Address))
	require.NoError(t, err)
	require.Equal(t, []types.RequestID{42}, k.GetPendingResolveList(ctx))
	event = abci.Event{
		Type: types.EventTypeReport,
		Attributes: []abci.EventAttribute{
			{Key: []byte(types.AttributeKeyID), Value: []byte("42")},
			{Key: []byte(types.AttributeKeyValidator), Value: []byte(testapp.Validators[1].ValAddress.String())},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
	// Even if we resolve the request, Validators[2] should still be able to report.
	k.SetPendingResolveList(ctx, []types.RequestID{})
	k.ResolveSuccess(ctx, 42, []byte("RESOLVE_RESULT!"), 1234)
	res, err = oracle.NewHandler(k)(ctx, types.NewMsgReportData(42, reports, testapp.Validators[2].ValAddress, testapp.Validators[2].Address))
	require.NoError(t, err)
	event = abci.Event{
		Type: types.EventTypeReport,
		Attributes: []abci.EventAttribute{
			{Key: []byte(types.AttributeKeyID), Value: []byte("42")},
			{Key: []byte(types.AttributeKeyValidator), Value: []byte(testapp.Validators[2].ValAddress.String())},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
	// Check the reports of this request. We should see 3 reports, with report from Validators[2] comes after resolve.
	finalReport := k.GetReports(ctx, 42)
	require.Contains(t, finalReport, types.NewReport(testapp.Validators[0].ValAddress, true, reports))
	require.Contains(t, finalReport, types.NewReport(testapp.Validators[1].ValAddress, true, reports))
	require.Contains(t, finalReport, types.NewReport(testapp.Validators[2].ValAddress, false, reports))
}

func TestReportFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	// Set up a mock request asking 3 validators with min count 2.
	k.SetRequest(ctx, 42, types.NewRequest(
		1,
		[]byte("beeb"),
		[]sdk.ValAddress{testapp.Validators[2].ValAddress, testapp.Validators[1].ValAddress, testapp.Validators[0].ValAddress},
		2,
		124,
		testapp.ParseTime(1581589790),
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
	res, err := oracle.NewHandler(k)(ctx, types.NewMsgReportData(999, reports, testapp.Validators[0].ValAddress, testapp.Validators[0].Address))
	testapp.CheckErrorf(t, err, types.ErrRequestNotFound, "id: 999")
	require.Nil(t, res)
	// Not-asked validator
	res, err = oracle.NewHandler(k)(ctx, types.NewMsgReportData(42, reports, testapp.Alice.ValAddress, testapp.Alice.Address))
	testapp.CheckErrorf(t, err, types.ErrValidatorNotRequested, "reqID: 42, val: %s", testapp.Alice.ValAddress.String())
	require.Nil(t, res)
	// Too large report data size
	res, err = oracle.NewHandler(k)(ctx, types.NewMsgReportData(42, []types.RawReport{types.NewRawReport(1, 0, []byte("data1")), types.NewRawReport(2, 0, []byte(strings.Repeat("data2", 2000)))}, testapp.Validators[0].ValAddress, testapp.Validators[0].Address))
	testapp.CheckErrorf(t, err, types.ErrTooLargeRawReportData, "got: 10000, max: 512")
	require.Nil(t, res)
	// Not an authorized reporter
	res, err = oracle.NewHandler(k)(ctx, types.NewMsgReportData(42, reports, testapp.Validators[0].ValAddress, testapp.Alice.Address))
	require.ErrorIs(t, err, types.ErrReporterNotAuthorized)
	require.Nil(t, res)
	// Not having all raw reports
	res, err = oracle.NewHandler(k)(ctx, types.NewMsgReportData(42, []types.RawReport{types.NewRawReport(1, 0, []byte("data1"))}, testapp.Validators[0].ValAddress, testapp.Validators[0].Address))
	require.ErrorIs(t, err, types.ErrInvalidReportSize)
	require.Nil(t, res)
	// Incorrect external IDs
	res, err = oracle.NewHandler(k)(ctx, types.NewMsgReportData(42, []types.RawReport{types.NewRawReport(1, 0, []byte("data1")), types.NewRawReport(42, 0, []byte("data2"))}, testapp.Validators[0].ValAddress, testapp.Validators[0].Address))
	testapp.CheckErrorf(t, err, types.ErrRawRequestNotFound, "reqID: 42, extID: 42")
	require.Nil(t, res)
	// Request already expired
	k.SetRequestLastExpired(ctx, 42)
	res, err = oracle.NewHandler(k)(ctx, types.NewMsgReportData(42, reports, testapp.Validators[0].ValAddress, testapp.Validators[0].Address))
	require.ErrorIs(t, err, types.ErrRequestAlreadyExpired)
	require.Nil(t, res)
}

func TestActivateSuccess(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1000000))
	require.Equal(t,
		types.NewValidatorStatus(false, time.Time{}),
		k.GetValidatorStatus(ctx, testapp.Validators[0].ValAddress),
	)
	msg := types.NewMsgActivate(testapp.Validators[0].ValAddress)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	require.Equal(t,
		types.NewValidatorStatus(true, testapp.ParseTime(1000000)),
		k.GetValidatorStatus(ctx, testapp.Validators[0].ValAddress),
	)
	event := abci.Event{
		Type: types.EventTypeActivate,
		Attributes: []abci.EventAttribute{
			{Key: []byte(types.AttributeKeyValidator), Value: []byte(testapp.Validators[0].ValAddress.String())},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
}

func TestActivateFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	msg := types.NewMsgActivate(testapp.Validators[0].ValAddress)
	// Already active.
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.ErrorIs(t, err, types.ErrValidatorAlreadyActive)
	require.Nil(t, res)
	// Too soon to activate.
	ctx = ctx.WithBlockTime(testapp.ParseTime(100000))
	k.MissReport(ctx, testapp.Validators[0].ValAddress, testapp.ParseTime(99999))
	ctx = ctx.WithBlockTime(testapp.ParseTime(100001))
	res, err = oracle.NewHandler(k)(ctx, msg)
	require.ErrorIs(t, err, types.ErrTooSoonToActivate)
	require.Nil(t, res)
	// OK
	ctx = ctx.WithBlockTime(testapp.ParseTime(200000))
	_, err = oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
}

func TestAddReporterSuccess(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	require.False(t, k.IsReporter(ctx, testapp.Alice.ValAddress, testapp.Bob.Address))
	// Add testapp.Bob to a reporter of testapp.Alice validator.
	msg := types.NewMsgAddReporter(testapp.Alice.ValAddress, testapp.Bob.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	require.True(t, k.IsReporter(ctx, testapp.Alice.ValAddress, testapp.Bob.Address))
	event := abci.Event{
		Type: types.EventTypeAddReporter,
		Attributes: []abci.EventAttribute{
			{Key: []byte(types.AttributeKeyValidator), Value: []byte(testapp.Alice.ValAddress.String())},
			{Key: []byte(types.AttributeKeyReporter), Value: []byte(testapp.Bob.Address.String())},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
}

func TestAddReporterFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	// Should fail when you try to add yourself as your reporter.
	msg := types.NewMsgAddReporter(testapp.Alice.ValAddress, testapp.Alice.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	testapp.CheckErrorf(t, err, types.ErrReporterAlreadyExists, "val: %s, addr: %s", testapp.Alice.ValAddress.String(), testapp.Alice.Address.String())
	require.Nil(t, res)
}

func TestRemoveReporterSuccess(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	// Add testapp.Bob to a reporter of testapp.Alice validator.
	err := k.AddReporter(ctx, testapp.Alice.ValAddress, testapp.Bob.Address)
	require.True(t, k.IsReporter(ctx, testapp.Alice.ValAddress, testapp.Bob.Address))
	require.NoError(t, err)
	// Now remove testapp.Bob from the set of testapp.Alice's reporters.
	msg := types.NewMsgRemoveReporter(testapp.Alice.ValAddress, testapp.Bob.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	require.False(t, k.IsReporter(ctx, testapp.Alice.ValAddress, testapp.Bob.Address))
	event := abci.Event{
		Type: types.EventTypeRemoveReporter,
		Attributes: []abci.EventAttribute{
			{Key: []byte(types.AttributeKeyValidator), Value: []byte(testapp.Alice.ValAddress.String())},
			{Key: []byte(types.AttributeKeyReporter), Value: []byte(testapp.Bob.Address.String())},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
}

func TestRemoveReporterFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	// Should fail because testapp.Bob isn't testapp.Alice validator's reporter.
	msg := types.NewMsgRemoveReporter(testapp.Alice.ValAddress, testapp.Bob.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	testapp.CheckErrorf(t, err, types.ErrReporterNotFound, "val: %s, addr: %s", testapp.Alice.ValAddress.String(), testapp.Bob.Address.String())
	require.Nil(t, res)
}
