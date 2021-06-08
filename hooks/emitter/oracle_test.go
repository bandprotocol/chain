package emitter

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/hooks/common"
	"github.com/bandprotocol/chain/testing/testapp"
	"github.com/bandprotocol/chain/x/oracle/types"
)

var (
	Calldata = []byte("Calldata")
)

func TestDecodeMsgRequestData(t *testing.T) {
	msgJson := make(common.JsDict)
	msg := types.NewMsgRequestData(1, []byte("calldata"), 1, 1, "cleint_id", testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	decodeMsgRequestData(msg, msgJson)
	require.Equal(t, msgJson, common.JsDict{
		"oracle_script_id": types.OracleScriptID(1),
		"calldata":         []byte("calldata"),
		"ask_count":        uint64(1),
		"min_count":        uint64(1),
		"client_id":        "cleint_id",
		"fee_limit":        testapp.Coins100000000uband,
		"prepare_gas":      testapp.TestDefaultPrepareGas,
		"execute_gas":      testapp.TestDefaultExecuteGas,
		"sender":           testapp.Alice.Address.String(),
	})
}

func TestDecodeReportData(t *testing.T) {
	msgJson := make(common.JsDict)
	msg := types.NewMsgReportData(1, []types.RawReport{{1, 1, []byte("data1")}, {2, 2, []byte("data2")}}, testapp.Validators[0].ValAddress, testapp.Alice.Address)
	decodeMsgReportData(msg, msgJson)
	require.Equal(t, msgJson, common.JsDict{
		"request_id":  types.RequestID(1),
		"raw_reports": []types.RawReport{{1, 1, []byte("data1")}, {2, 2, []byte("data2")}},
		"validator":   testapp.Validators[0].ValAddress.String(),
		"reporter":    testapp.Alice.Address.String(),
	})
}

func TestDecodeMsgCreateDataSource(t *testing.T) {
	msgJson := make(common.JsDict)
	msg := types.NewMsgCreateDataSource("name", "desc", []byte("exec"), testapp.Coins1000000uband, testapp.Treasury.Address, testapp.Owner.Address, testapp.Alice.Address)
	decodeMsgCreateDataSource(msg, msgJson)
	require.Equal(t, msgJson, common.JsDict{
		"name":        "name",
		"description": "desc",
		"executable":  []byte("exec"),
		"fee":         testapp.Coins1000000uband,
		"treasury":    testapp.Treasury.Address.String(),
		"owner":       testapp.Owner.Address.String(),
		"sender":      testapp.Alice.Address.String(),
	})
}

func TestDecodeCreateOracleScript(t *testing.T) {
	msgJson := make(common.JsDict)
	msg := types.NewMsgCreateOracleScript("name", "desc", "schema", "url", []byte("code"), testapp.Owner.Address, testapp.Alice.Address)
	decodeMsgCreateOracleScript(msg, msgJson)
	require.Equal(t, msgJson, common.JsDict{
		"name":            "name",
		"description":     "desc",
		"schema":          "schema",
		"source_code_url": "url",
		"code":            []byte("code"),
		"owner":           testapp.Owner.Address.String(),
		"sender":          testapp.Alice.Address.String(),
	})
}

func TestDecodeMsgEditDataSource(t *testing.T) {
	msgJson := make(common.JsDict)
	msg := types.NewMsgEditDataSource(1, "name", "desc", []byte("exec"), testapp.Coins1000000uband, testapp.Treasury.Address, testapp.Owner.Address, testapp.Alice.Address)
	decodeMsgEditDataSource(msg, msgJson)
	require.Equal(t, msgJson, common.JsDict{
		"data_source_id": types.DataSourceID(1),
		"name":           "name",
		"description":    "desc",
		"executable":     []byte("exec"),
		"fee":            testapp.Coins1000000uband,
		"treasury":       testapp.Treasury.Address.String(),
		"owner":          testapp.Owner.Address.String(),
		"sender":         testapp.Alice.Address.String(),
	})
}

func TestDecodeMsgEditOracleScript(t *testing.T) {
	msgJson := make(common.JsDict)
	msg := types.NewMsgEditOracleScript(1, "name", "desc", "schema", "url", []byte("code"), testapp.Owner.Address, testapp.Alice.Address)
	decodeMsgEditOracleScript(msg, msgJson)
	require.Equal(t, msgJson, common.JsDict{
		"oracle_script_id": types.OracleScriptID(1),
		"name":             "name",
		"description":      "desc",
		"schema":           "schema",
		"source_code_url":  "url",
		"code":             []byte("code"),
		"owner":            testapp.Owner.Address.String(),
		"sender":           testapp.Alice.Address.String(),
	})
}

func TestDecodeMsgAddReporter(t *testing.T) {
	msgJson := make(common.JsDict)
	msg := types.NewMsgAddReporter(testapp.Alice.ValAddress, testapp.Bob.Address)
	decodeMsgAddReporter(msg, msgJson)
	require.Equal(t, msgJson, common.JsDict{
		"validator": testapp.Alice.ValAddress.String(),
		"reporter":  testapp.Bob.Address.String(),
	})
}

func TestDecodeMsgRemoveReporter(t *testing.T) {
	msgJson := make(common.JsDict)
	msg := types.NewMsgRemoveReporter(testapp.Alice.ValAddress, testapp.Bob.Address)
	decodeMsgRemoveReporter(msg, msgJson)
	require.Equal(t, msgJson, common.JsDict{
		"validator": testapp.Alice.ValAddress.String(),
		"reporter":  testapp.Bob.Address.String(),
	})
}

func TestDecodeMsgActivate(t *testing.T) {
	msgJson := make(common.JsDict)
	msg := types.NewMsgActivate(testapp.Alice.ValAddress)
	decodeMsgActivate(msg, msgJson)
	require.Equal(t, msgJson, common.JsDict{
		"validator": testapp.Alice.ValAddress.String(),
	})
}
