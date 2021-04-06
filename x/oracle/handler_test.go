package oracle_test

import (
	"bytes"
	gz "compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/GeoDB-Limited/odin-core/x/common/testapp"
	"github.com/GeoDB-Limited/odin-core/x/oracle"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
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
	msg := oracletypes.NewMsgCreateDataSource(name, description, executable, testapp.EmptyCoins, treasury, owner, testapp.Alice.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	ds, err := k.GetDataSource(ctx, oracletypes.DataSourceID(dsCount+1))
	require.NoError(t, err)
	require.Equal(t, oracletypes.NewDataSource(testapp.Owner.Address, name, description, filename, testapp.Treasury.Address, testapp.EmptyCoins), ds)
	event := abci.Event{
		Type:       oracletypes.EventTypeCreateDataSource,
		Attributes: []abci.EventAttribute{{Key: []byte(oracletypes.AttributeKeyID), Value: []byte(fmt.Sprintf("%d", dsCount+1))}},
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
	msg := oracletypes.NewMsgCreateDataSource(name, description, buf.Bytes()[:5], testapp.EmptyCoins, treasury, owner, sender)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.EqualError(t, err, "unexpected EOF: uncompression failed")
	require.Nil(t, res)
}

func TestEditDataSourceSuccess(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	newName := "beeb"
	newDescription := "new_description"
	newExecutable := []byte("executable2")
	newExecutableHash := sha256.Sum256(newExecutable)
	newFilename := hex.EncodeToString(newExecutableHash[:])
	msg := oracletypes.NewMsgEditDataSource(1, newName, newDescription, newExecutable, testapp.EmptyCoins, testapp.Treasury.Address, testapp.Owner.Address, testapp.Owner.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	ds, err := k.GetDataSource(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, oracletypes.NewDataSource(testapp.Owner.Address, newName, newDescription, newFilename, testapp.Treasury.Address, testapp.Coins1000000odin), ds)
	event := abci.Event{
		Type:       oracletypes.EventTypeEditDataSource,
		Attributes: []abci.EventAttribute{{Key: []byte(oracletypes.AttributeKeyID), Value: []byte("1")}},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
}

func TestEditDataSourceFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	newName := "beeb"
	newDescription := "new_description"
	newExecutable := []byte("executable2")
	// Bad ID
	msg := oracletypes.NewMsgEditDataSource(42, newName, newDescription, newExecutable, testapp.EmptyCoins, testapp.Treasury.Address, testapp.Owner.Address, testapp.Owner.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	// require.EqualError(t, err, "data source not found: id: 42")
	require.Nil(t, res)
	// Not owner
	msg = oracletypes.NewMsgEditDataSource(1, newName, newDescription, newExecutable, testapp.EmptyCoins, testapp.Treasury.Address, testapp.Owner.Address, testapp.Bob.Address)
	res, err = oracle.NewHandler(k)(ctx, msg)
	// require.EqualError(t, err, "editor not authorized")
	require.Nil(t, res)
	// Bad Gzip
	var buf bytes.Buffer
	zw := gz.NewWriter(&buf)
	zw.Write(newExecutable)
	zw.Close()
	msg = oracletypes.NewMsgEditDataSource(1, newName, newDescription, buf.Bytes()[:5], testapp.EmptyCoins, testapp.Treasury.Address, testapp.Owner.Address, testapp.Owner.Address)
	res, err = oracle.NewHandler(k)(ctx, msg)
	// require.EqualError(t, err, "uncompression failed: unexpected EOF")
	_ = err
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
	msg := oracletypes.NewMsgCreateOracleScript(name, description, schema, url, code, testapp.Owner.Address, testapp.Alice.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	os, err := k.GetOracleScript(ctx, oracletypes.OracleScriptID(osCount+1))
	require.NoError(t, err)
	require.Equal(t, oracletypes.NewOracleScript(testapp.Owner.Address, name, description, testapp.WasmExtra1FileName, schema, url), os)

	event := abci.Event{
		Type:       oracletypes.EventTypeCreateOracleScript,
		Attributes: []abci.EventAttribute{{Key: []byte(oracletypes.AttributeKeyID), Value: []byte(fmt.Sprintf("%d", osCount+1))}},
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
	msg := oracletypes.NewMsgCreateOracleScript(name, description, schema, url, buf.Bytes(), testapp.Owner.Address, testapp.Alice.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	os, err := k.GetOracleScript(ctx, oracletypes.OracleScriptID(osCount+1))
	require.NoError(t, err)
	require.Equal(t, oracletypes.NewOracleScript(testapp.Owner.Address, name, description, testapp.WasmExtra1FileName, schema, url), os)

	event := abci.Event{
		Type:       oracletypes.EventTypeCreateOracleScript,
		Attributes: []abci.EventAttribute{{Key: []byte(oracletypes.AttributeKeyID), Value: []byte(fmt.Sprintf("%d", osCount+1))}},
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
	msg := oracletypes.NewMsgCreateOracleScript(name, description, schema, url, []byte("BAD"), testapp.Owner.Address, testapp.Alice.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	// require.EqualError(t, err, "owasm compilation failed: with error: wasm code does not pass basic validation")
	require.Nil(t, res)
	// Bad Gzip
	var buf bytes.Buffer
	zw := gz.NewWriter(&buf)
	zw.Write(testapp.WasmExtra1)
	zw.Close()
	msg = oracletypes.NewMsgCreateOracleScript(name, description, schema, url, buf.Bytes()[:5], testapp.Owner.Address, testapp.Alice.Address)
	res, err = oracle.NewHandler(k)(ctx, msg)
	// require.EqualError(t, err, "uncompression failed: unexpected EOF")
	_ = err
	require.Nil(t, res)
}

func TestEditOracleScriptSuccess(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	newName := "os_2"
	newDescription := "beebbeeb"
	newCode := testapp.WasmExtra2
	newSchema := "new_schema"
	newURL := "new_url"
	msg := oracletypes.NewMsgEditOracleScript(1, newName, newDescription, newSchema, newURL, newCode, testapp.Owner.Address, testapp.Owner.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	os, err := k.GetOracleScript(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, oracletypes.NewOracleScript(testapp.Owner.Address, newName, newDescription, testapp.WasmExtra2FileName, newSchema, newURL), os)

	event := abci.Event{
		Type:       oracletypes.EventTypeEditOracleScript,
		Attributes: []abci.EventAttribute{{Key: []byte(oracletypes.AttributeKeyID), Value: []byte("1")}},
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
	msg := oracletypes.NewMsgEditOracleScript(999, newName, newDescription, newSchema, newURL, newCode, testapp.Owner.Address, testapp.Owner.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	// require.EqualError(t, err, "oracle script not found: id: 999")
	require.Nil(t, res)
	// Not owner
	msg = oracletypes.NewMsgEditOracleScript(1, newName, newDescription, newSchema, newURL, newCode, testapp.Owner.Address, testapp.Bob.Address)
	res, err = oracle.NewHandler(k)(ctx, msg)
	require.EqualError(t, err, "editor not authorized")
	require.Nil(t, res)
	// Bad Owasm code
	msg = oracletypes.NewMsgEditOracleScript(1, newName, newDescription, newSchema, newURL, []byte("BAD_CODE"), testapp.Owner.Address, testapp.Owner.Address)
	res, err = oracle.NewHandler(k)(ctx, msg)
	// require.EqualError(t, err, "owasm compilation failed: with error: wasm code does not pass basic validation")
	require.Nil(t, res)
	// Bad Gzip
	var buf bytes.Buffer
	zw := gz.NewWriter(&buf)
	zw.Write(testapp.WasmExtra2)
	zw.Close()
	msg = oracletypes.NewMsgEditOracleScript(1, newName, newDescription, newSchema, newURL, buf.Bytes()[:5], testapp.Owner.Address, testapp.Owner.Address)
	res, err = oracle.NewHandler(k)(ctx, msg)
	// require.EqualError(t, err, "uncompression failed: unexpected EOF")
	_ = err
	require.Nil(t, res)
}

func TestRequestDataSuccess(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockHeight(124).WithBlockTime(testapp.ParseTime(1581589790))
	msg := oracletypes.NewMsgRequestData(1, []byte("beeb"), 2, 2, "CID", testapp.Coins10000000000odin, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.FeePayer.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, oracletypes.NewRequest(
		1,
		[]byte("beeb"),
		[]sdk.ValAddress{testapp.Validators[2].ValAddress, testapp.Validators[0].ValAddress},
		2,
		124,
		testapp.ParseTime(1581589790),
		"CID",
		[]oracletypes.RawRequest{
			oracletypes.NewRawRequest(1, 1, []byte("beeb")),
			oracletypes.NewRawRequest(2, 2, []byte("beeb")),
			oracletypes.NewRawRequest(3, 3, []byte("beeb")),
		},
		nil,
		uint64(oracletypes.DefaultExecuteGas),
	), k.MustGetRequest(ctx, 1))

	event := abci.Event{
		Type: banktypes.EventTypeTransfer,
		Attributes: []abci.EventAttribute{
			{Key: []byte(banktypes.AttributeKeyRecipient), Value: []byte(testapp.Treasury.Address.String())},
			{Key: []byte(banktypes.AttributeKeySender), Value: []byte(testapp.FeePayer.Address.String())},
			{Key: []byte(sdk.AttributeKeyAmount), Value: []byte("2000000odin")},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
	require.Equal(t, abci.Event(event), res.Events[2])
	require.Equal(t, abci.Event(event), res.Events[4])
	event = abci.Event{
		Type: sdk.EventTypeMessage,
		Attributes: []abci.EventAttribute{
			{Key: []byte(banktypes.AttributeKeySender), Value: []byte(testapp.FeePayer.Address.String())},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[1])
	require.Equal(t, abci.Event(event), res.Events[3])
	require.Equal(t, abci.Event(event), res.Events[5])

	event = abci.Event{
		Type: oracletypes.EventTypeRequest,
		Attributes: []abci.EventAttribute{
			{Key: []byte(oracletypes.AttributeKeyID), Value: []byte("1")},
			{Key: []byte(oracletypes.AttributeKeyClientID), Value: []byte("CID")},
			{Key: []byte(oracletypes.AttributeKeyOracleScriptID), Value: []byte("1")},
			{Key: []byte(oracletypes.AttributeKeyCalldata), Value: []byte("62656562")}, // "beeb" in hex
			{Key: []byte(oracletypes.AttributeKeyAskCount), Value: []byte("2")},
			{Key: []byte(oracletypes.AttributeKeyMinCount), Value: []byte("2")},
			{Key: []byte(oracletypes.AttributeKeyGasUsed), Value: []byte("3089")},
			{Key: []byte(oracletypes.AttributeKeyValidator), Value: []byte(testapp.Validators[2].ValAddress.String())},
			{Key: []byte(oracletypes.AttributeKeyValidator), Value: []byte(testapp.Validators[0].ValAddress.String())},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[6])
	event = abci.Event{
		Type: oracletypes.EventTypeRawRequest,
		Attributes: []abci.EventAttribute{
			{Key: []byte(oracletypes.AttributeKeyDataSourceID), Value: []byte("1")},
			{Key: []byte(oracletypes.AttributeKeyDataSourceHash), Value: []byte(testapp.DataSources[1].Filename)},
			{Key: []byte(oracletypes.AttributeKeyExternalID), Value: []byte("1")},
			{Key: []byte(oracletypes.AttributeKeyCalldata), Value: []byte("beeb")},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[7])
	event = abci.Event{
		Type: oracletypes.EventTypeRawRequest,
		Attributes: []abci.EventAttribute{
			{Key: []byte(oracletypes.AttributeKeyDataSourceID), Value: []byte("2")},
			{Key: []byte(oracletypes.AttributeKeyDataSourceHash), Value: []byte(testapp.DataSources[2].Filename)},
			{Key: []byte(oracletypes.AttributeKeyExternalID), Value: []byte("2")},
			{Key: []byte(oracletypes.AttributeKeyCalldata), Value: []byte("beeb")},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[8])
	event = abci.Event{
		Type: oracletypes.EventTypeRawRequest,
		Attributes: []abci.EventAttribute{
			{Key: []byte(oracletypes.AttributeKeyDataSourceID), Value: []byte("3")},
			{Key: []byte(oracletypes.AttributeKeyDataSourceHash), Value: []byte(testapp.DataSources[3].Filename)},
			{Key: []byte(oracletypes.AttributeKeyExternalID), Value: []byte("3")},
			{Key: []byte(oracletypes.AttributeKeyCalldata), Value: []byte("beeb")},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[9])
}

func TestRequestDataFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	// No active oracle validators
	res, err := oracle.NewHandler(k)(ctx, oracletypes.NewMsgRequestData(1, []byte("beeb"), 2, 2, "CID", testapp.Coins100000000odin, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.FeePayer.Address))
	require.EqualError(t, err, "0 < 2: insufficent available validators")
	require.Nil(t, res)
	k.Activate(ctx, testapp.Validators[0].ValAddress)
	k.Activate(ctx, testapp.Validators[1].ValAddress)
	// Too high ask count
	res, err = oracle.NewHandler(k)(ctx, oracletypes.NewMsgRequestData(1, []byte("beeb"), 3, 2, "CID", testapp.Coins100000000odin, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.FeePayer.Address))
	require.EqualError(t, err, "2 < 3: insufficent available validators")
	require.Nil(t, res)
	// Bad oracle script ID
	res, err = oracle.NewHandler(k)(ctx, oracletypes.NewMsgRequestData(999, []byte("beeb"), 2, 2, "CID", testapp.Coins100000000odin, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.FeePayer.Address))
	require.EqualError(t, err, "id: 999: oracle script not found")
	require.Nil(t, res)
	// Pay not enough fee
	res, err = oracle.NewHandler(k)(ctx, oracletypes.NewMsgRequestData(1, []byte("beeb"), 2, 2, "CID", testapp.EmptyCoins, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.FeePayer.Address))
	require.EqualError(t, err, "require: 2000000odin, max: 0odin: not enough fee")
	require.Nil(t, res)
	// Too large calldata
	res, err = oracle.NewHandler(k)(ctx, oracletypes.NewMsgRequestData(999, []byte(strings.Repeat("a", 2000)), 2, 2, "CID", testapp.Coins100000000odin, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.Alice.Address))
	require.EqualError(t, err, "got: 2000, max: 1024: too large calldata")
	require.Nil(t, res)
}

func TestReportSuccess(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	// Set up a mock request asking 3 validators with min count 2.
	k.SetRequest(ctx, 42, oracletypes.NewRequest(
		1,
		[]byte("beeb"),
		[]sdk.ValAddress{testapp.Validators[2].ValAddress, testapp.Validators[1].ValAddress, testapp.Validators[0].ValAddress},
		2,
		124,
		testapp.ParseTime(1581589790),
		"CID",
		[]oracletypes.RawRequest{
			oracletypes.NewRawRequest(1, 1, []byte("beeb")),
			oracletypes.NewRawRequest(2, 2, []byte("beeb")),
		},
		nil,
		0,
	))
	// Common raw reports for everyone.
	reports := []oracletypes.RawReport{oracletypes.NewRawReport(1, 0, []byte("data1")), oracletypes.NewRawReport(2, 0, []byte("data2"))}
	// Validators[0] reports data.
	res, err := oracle.NewHandler(k)(ctx, oracletypes.NewMsgReportData(42, reports, testapp.Validators[0].ValAddress, testapp.Validators[0].Address))
	require.NoError(t, err)
	require.Equal(t, []oracletypes.RequestID{}, k.GetPendingResolveList(ctx))
	event := abci.Event{
		Type: oracletypes.EventTypeReport,
		Attributes: []abci.EventAttribute{
			{Key: []byte(oracletypes.AttributeKeyID), Value: []byte("42")},
			{Key: []byte(oracletypes.AttributeKeyValidator), Value: []byte(testapp.Validators[0].ValAddress.String())},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
	// Validators[1] reports data. Now the request should move to pending resolve.
	res, err = oracle.NewHandler(k)(ctx, oracletypes.NewMsgReportData(42, reports, testapp.Validators[1].ValAddress, testapp.Validators[1].Address))
	require.NoError(t, err)
	require.Equal(t, []oracletypes.RequestID{42}, k.GetPendingResolveList(ctx))
	event = abci.Event{
		Type: oracletypes.EventTypeReport,
		Attributes: []abci.EventAttribute{
			{Key: []byte(oracletypes.AttributeKeyID), Value: []byte("42")},
			{Key: []byte(oracletypes.AttributeKeyValidator), Value: []byte(testapp.Validators[1].ValAddress.String())},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
	// Even if we resolve the request, Validators[2] should still be able to report.
	k.SetPendingResolveList(ctx, []oracletypes.RequestID{})
	k.ResolveSuccess(ctx, 42, []byte("RESOLVE_RESULT!"), 1234)
	res, err = oracle.NewHandler(k)(ctx, oracletypes.NewMsgReportData(42, reports, testapp.Validators[2].ValAddress, testapp.Validators[2].Address))
	require.NoError(t, err)
	event = abci.Event{
		Type: oracletypes.EventTypeReport,
		Attributes: []abci.EventAttribute{
			{Key: []byte(oracletypes.AttributeKeyID), Value: []byte("42")},
			{Key: []byte(oracletypes.AttributeKeyValidator), Value: []byte(testapp.Validators[2].ValAddress.String())},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
	// require.Equal(t, k.GetReports(ctx, 42)[0], oracletypes.NewReport(testapp.Validators[0].ValAddress, false, reports))
	// Check the reports of this request. We should see 3 reports, with report from Validators[2] comes after resolve.
	require.Contains(t, k.GetReports(ctx, 42), oracletypes.NewReport(testapp.Validators[0].ValAddress, true, reports))
	require.Contains(t, k.GetReports(ctx, 42), oracletypes.NewReport(testapp.Validators[1].ValAddress, true, reports))
	require.Contains(t, k.GetReports(ctx, 42), oracletypes.NewReport(testapp.Validators[2].ValAddress, false, reports))
}

func TestReportFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	// Set up a mock request asking 3 validators with min count 2.
	k.SetRequest(ctx, 42, oracletypes.NewRequest(
		1,
		[]byte("beeb"),
		[]sdk.ValAddress{testapp.Validators[2].ValAddress, testapp.Validators[1].ValAddress, testapp.Validators[0].ValAddress},
		2,
		124,
		testapp.ParseTime(1581589790),
		"CID",
		[]oracletypes.RawRequest{
			oracletypes.NewRawRequest(1, 1, []byte("beeb")),
			oracletypes.NewRawRequest(2, 2, []byte("beeb")),
		},
		nil,
		0,
	))
	// Common raw reports for everyone.
	reports := []oracletypes.RawReport{oracletypes.NewRawReport(1, 0, []byte("data1")), oracletypes.NewRawReport(2, 0, []byte("data2"))}
	// Bad ID
	res, err := oracle.NewHandler(k)(ctx, oracletypes.NewMsgReportData(999, reports, testapp.Validators[0].ValAddress, testapp.Validators[0].Address))
	// require.EqualError(t, err, "request not found: id: 999")
	require.Nil(t, res)
	// Not-asked validator
	res, err = oracle.NewHandler(k)(ctx, oracletypes.NewMsgReportData(42, reports, testapp.Alice.ValAddress, testapp.Alice.Address))
	// require.EqualError(t, err, fmt.Sprintf("validator not requested: reqID: 42, val: %s", testapp.Alice.ValAddress.String()))
	require.Nil(t, res)
	// Not an authorized reporter
	res, err = oracle.NewHandler(k)(ctx, oracletypes.NewMsgReportData(42, reports, testapp.Validators[0].ValAddress, testapp.Alice.Address))
	// require.EqualError(t, err, "reporter not authorized")
	require.Nil(t, res)
	// Not having all raw reports
	res, err = oracle.NewHandler(k)(ctx, oracletypes.NewMsgReportData(42, []oracletypes.RawReport{oracletypes.NewRawReport(1, 0, []byte("data1"))}, testapp.Validators[0].ValAddress, testapp.Validators[0].Address))
	// require.EqualError(t, err, "invalid report size")
	require.Nil(t, res)
	// Incorrect external IDs
	res, err = oracle.NewHandler(k)(ctx, oracletypes.NewMsgReportData(42, []oracletypes.RawReport{oracletypes.NewRawReport(1, 0, []byte("data1")), oracletypes.NewRawReport(42, 0, []byte("data2"))}, testapp.Validators[0].ValAddress, testapp.Validators[0].Address))
	// require.EqualError(t, err, "raw request not found: reqID: 42, extID: 42")
	require.Nil(t, res)
	// Request already expired
	k.SetRequestLastExpired(ctx, 42)
	res, err = oracle.NewHandler(k)(ctx, oracletypes.NewMsgReportData(42, reports, testapp.Validators[0].ValAddress, testapp.Validators[0].Address))
	// require.EqualError(t, err, "request already expired")
	require.Nil(t, res)
	_ = err
}

func TestActivateSuccess(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1000000))
	require.Equal(t,
		oracletypes.NewValidatorStatus(false, time.Time{}),
		k.GetValidatorStatus(ctx, testapp.Validators[0].ValAddress),
	)
	msg := oracletypes.NewMsgActivate(testapp.Validators[0].ValAddress)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	require.Equal(t,
		oracletypes.NewValidatorStatus(true, testapp.ParseTime(1000000)),
		k.GetValidatorStatus(ctx, testapp.Validators[0].ValAddress),
	)
	event := abci.Event{
		Type: oracletypes.EventTypeActivate,
		Attributes: []abci.EventAttribute{
			{Key: []byte(oracletypes.AttributeKeyValidator), Value: []byte(testapp.Validators[0].ValAddress.String())},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
}

func TestActivateFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	msg := oracletypes.NewMsgActivate(testapp.Validators[0].ValAddress)
	// Already active.
	res, err := oracle.NewHandler(k)(ctx, msg)
	// require.EqualError(t, err, "validator already active")
	require.Nil(t, res)
	// Too soon to activate.
	ctx = ctx.WithBlockTime(testapp.ParseTime(100000))
	k.MissReport(ctx, testapp.Validators[0].ValAddress, testapp.ParseTime(99999))
	ctx = ctx.WithBlockTime(testapp.ParseTime(100001))
	res, err = oracle.NewHandler(k)(ctx, msg)
	// require.EqualError(t, err, "too soon to activate")
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
	msg := oracletypes.NewMsgAddReporter(testapp.Alice.ValAddress, testapp.Bob.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	require.True(t, k.IsReporter(ctx, testapp.Alice.ValAddress, testapp.Bob.Address))
	event := abci.Event{
		Type: oracletypes.EventTypeAddReporter,
		Attributes: []abci.EventAttribute{
			{Key: []byte(oracletypes.AttributeKeyValidator), Value: []byte(testapp.Alice.ValAddress.String())},
			{Key: []byte(oracletypes.AttributeKeyReporter), Value: []byte(testapp.Bob.Address.String())},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
}

func TestAddReporterFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	// Should fail when you try to add yourself as your reporter.
	msg := oracletypes.NewMsgAddReporter(testapp.Alice.ValAddress, testapp.Alice.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.EqualError(t, err, fmt.Sprintf("val: %s, addr: %s: reporter already exists", testapp.Alice.ValAddress.String(), testapp.Alice.Address.String()))
	_ = err
	require.Nil(t, res)
	// Too large report data
	k.SetRequestLastExpired(ctx, 0)
	reports := []oracletypes.RawReport{oracletypes.NewRawReport(1, 0, []byte(strings.Repeat("1", 2000))), oracletypes.NewRawReport(2, 0, []byte(strings.Repeat("2", 2000)))}
	res, err = oracle.NewHandler(k)(ctx, oracletypes.NewMsgReportData(1, reports, testapp.Validators[0].ValAddress, testapp.Validators[0].Address))
	require.EqualError(t, err, "got: 2000, max: 1024: too large raw report data")
	require.Nil(t, res)
}

func TestRemoveReporterSuccess(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	// Add testapp.Bob to a reporter of testapp.Alice validator.
	err := k.AddReporter(ctx, testapp.Alice.ValAddress, testapp.Bob.Address)
	require.True(t, k.IsReporter(ctx, testapp.Alice.ValAddress, testapp.Bob.Address))
	require.NoError(t, err)
	// Now remove testapp.Bob from the set of testapp.Alice's reporters.
	msg := oracletypes.NewMsgRemoveReporter(testapp.Alice.ValAddress, testapp.Bob.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	require.NoError(t, err)
	require.False(t, k.IsReporter(ctx, testapp.Alice.ValAddress, testapp.Bob.Address))
	event := abci.Event{
		Type: oracletypes.EventTypeRemoveReporter,
		Attributes: []abci.EventAttribute{
			{Key: []byte(oracletypes.AttributeKeyValidator), Value: []byte(testapp.Alice.ValAddress.String())},
			{Key: []byte(oracletypes.AttributeKeyReporter), Value: []byte(testapp.Bob.Address.String())},
		},
	}
	require.Equal(t, abci.Event(event), res.Events[0])
}

func TestRemoveReporterFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	// Should fail because testapp.Bob isn't testapp.Alice validator's reporter.
	msg := oracletypes.NewMsgRemoveReporter(testapp.Alice.ValAddress, testapp.Bob.Address)
	res, err := oracle.NewHandler(k)(ctx, msg)
	// require.EqualError(t, err, fmt.Sprintf("reporter not found: val: %s, addr: %s", testapp.Alice.ValAddress.String(), testapp.Bob.Address.String()))
	_ = err
	require.Nil(t, res)
}
