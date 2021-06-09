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
	msg := types.NewMsgRequestData(1, []byte("calldata"), 1, 1, "cleint_id", testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, SenderAddress)
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
		"sender":           SenderAddress.String(),
	})
	testCompareJson(t, msg,
		"{\"oracle_script_id\":1,\"calldata\":\"Y2FsbGRhdGE=\",\"ask_count\":1,\"min_count\":1,\"client_id\":\"cleint_id\",\"fee_limit\":[{\"denom\":\"uband\",\"amount\":\"100000000\"}],\"prepare_gas\":40000,\"execute_gas\":300000,\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}",
	)
}

func TestDecodeReportData(t *testing.T) {
	msgJson := make(common.JsDict)
	msg := types.NewMsgReportData(1, []types.RawReport{{1, 1, []byte("data1")}, {2, 2, []byte("data2")}}, ValAddress, ReporterAddress)
	decodeMsgReportData(msg, msgJson)
	require.Equal(t, msgJson, common.JsDict{
		"request_id":  types.RequestID(1),
		"raw_reports": []types.RawReport{{1, 1, []byte("data1")}, {2, 2, []byte("data2")}},
		"validator":   ValAddress.String(),
		"reporter":    ReporterAddress.String(),
	})
	testCompareJson(t, msg,
		"{\"request_id\":1,\"raw_reports\":[{\"external_id\":1,\"exit_code\":1,\"data\":\"ZGF0YTE=\"},{\"external_id\":2,\"exit_code\":2,\"data\":\"ZGF0YTI=\"}],\"validator\":\"bandvaloper1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqx6y767\",\"reporter\":\"band1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq2vqal4\"}",
	)
}

func TestDecodeMsgCreateDataSource(t *testing.T) {
	msgJson := make(common.JsDict)
	msg := types.NewMsgCreateDataSource("name", "desc", []byte("exec"), testapp.Coins1000000uband, TreasuryAddress, OwnerAddress, SenderAddress)
	decodeMsgCreateDataSource(msg, msgJson)
	require.Equal(t, msgJson, common.JsDict{
		"name":        "name",
		"description": "desc",
		"executable":  []byte("exec"),
		"fee":         testapp.Coins1000000uband,
		"treasury":    TreasuryAddress.String(),
		"owner":       OwnerAddress.String(),
		"sender":      SenderAddress.String(),
	})
	testCompareJson(t, msg,
		"{\"name\":\"name\",\"description\":\"desc\",\"executable\":\"ZXhlYw==\",\"fee\":[{\"denom\":\"uband\",\"amount\":\"1000000\"}],\"treasury\":\"band123ex2ctnw4e8jqqqqqqqqqqqqqqqqqqqrmzwp0\",\"owner\":\"band1famkuetjqqqqqqqqqqqqqqqqqqqqqqqqkzrxfg\",\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}",
	)
}

func TestDecodeCreateOracleScript(t *testing.T) {
	msgJson := make(common.JsDict)
	msg := types.NewMsgCreateOracleScript("name", "desc", "schema", "url", []byte("code"), OwnerAddress, SenderAddress)
	decodeMsgCreateOracleScript(msg, msgJson)
	require.Equal(t, msgJson, common.JsDict{
		"name":            "name",
		"description":     "desc",
		"schema":          "schema",
		"source_code_url": "url",
		"code":            []byte("code"),
		"owner":           OwnerAddress.String(),
		"sender":          SenderAddress.String(),
	})
	testCompareJson(t, msg,
		"{\"name\":\"name\",\"description\":\"desc\",\"schema\":\"schema\",\"source_code_url\":\"url\",\"code\":\"Y29kZQ==\",\"owner\":\"band1famkuetjqqqqqqqqqqqqqqqqqqqqqqqqkzrxfg\",\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}",
	)
}

func TestDecodeMsgEditDataSource(t *testing.T) {
	msgJson := make(common.JsDict)
	msg := types.NewMsgEditDataSource(1, "name", "desc", []byte("exec"), testapp.Coins1000000uband, TreasuryAddress, OwnerAddress, SenderAddress)
	decodeMsgEditDataSource(msg, msgJson)
	require.Equal(t, msgJson, common.JsDict{
		"data_source_id": types.DataSourceID(1),
		"name":           "name",
		"description":    "desc",
		"executable":     []byte("exec"),
		"fee":            testapp.Coins1000000uband,
		"treasury":       TreasuryAddress.String(),
		"owner":          OwnerAddress.String(),
		"sender":         SenderAddress.String(),
	})
	testCompareJson(t, msg,
		"{\"data_source_id\":1,\"name\":\"name\",\"description\":\"desc\",\"executable\":\"ZXhlYw==\",\"fee\":[{\"denom\":\"uband\",\"amount\":\"1000000\"}],\"treasury\":\"band123ex2ctnw4e8jqqqqqqqqqqqqqqqqqqqrmzwp0\",\"owner\":\"band1famkuetjqqqqqqqqqqqqqqqqqqqqqqqqkzrxfg\",\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}",
	)
}

func TestDecodeMsgEditOracleScript(t *testing.T) {
	msgJson := make(common.JsDict)
	msg := types.NewMsgEditOracleScript(1, "name", "desc", "schema", "url", []byte("code"), OwnerAddress, SenderAddress)
	decodeMsgEditOracleScript(msg, msgJson)
	require.Equal(t, msgJson, common.JsDict{
		"oracle_script_id": types.OracleScriptID(1),
		"name":             "name",
		"description":      "desc",
		"schema":           "schema",
		"source_code_url":  "url",
		"code":             []byte("code"),
		"owner":            OwnerAddress.String(),
		"sender":           SenderAddress.String(),
	})
	testCompareJson(t, msg,
		"{\"oracle_script_id\":1,\"name\":\"name\",\"description\":\"desc\",\"schema\":\"schema\",\"source_code_url\":\"url\",\"code\":\"Y29kZQ==\",\"owner\":\"band1famkuetjqqqqqqqqqqqqqqqqqqqqqqqqkzrxfg\",\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}",
	)
}

func TestDecodeMsgAddReporter(t *testing.T) {
	msgJson := make(common.JsDict)
	msg := types.NewMsgAddReporter(ValAddress, ReporterAddress)
	decodeMsgAddReporter(msg, msgJson)
	require.Equal(t, msgJson, common.JsDict{
		"validator": ValAddress.String(),
		"reporter":  ReporterAddress.String(),
	})
	testCompareJson(t, msg,
		"{\"validator\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\",\"reporter\":\"band12fjhqmmjw3jhyqqqqqqqqqqqqqqqqqqqjfy83g\"}",
	)
}

func TestDecodeMsgRemoveReporter(t *testing.T) {
	msgJson := make(common.JsDict)
	msg := types.NewMsgRemoveReporter(ValAddress, ReporterAddress)
	decodeMsgRemoveReporter(msg, msgJson)
	require.Equal(t, msgJson, common.JsDict{
		"validator": ValAddress.String(),
		"reporter":  ReporterAddress.String(),
	})
	testCompareJson(t, msg,
		"{\"validator\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\",\"reporter\":\"band12fjhqmmjw3jhyqqqqqqqqqqqqqqqqqqqjfy83g\"}",
	)
}

func TestDecodeMsgActivate(t *testing.T) {
	msgJson := make(common.JsDict)
	msg := types.NewMsgActivate(ValAddress)
	decodeMsgActivate(msg, msgJson)
	require.Equal(t, msgJson, common.JsDict{
		"validator": ValAddress.String(),
	})
	testCompareJson(t, msg,
		"{\"validator\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}
