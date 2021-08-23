package types

import (
	"encoding/hex"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestRequestStoreKey(t *testing.T) {
	expect, _ := hex.DecodeString("010000000000000014")
	require.Equal(t, expect, RequestStoreKey(20))
}

func TestReportStoreKey(t *testing.T) {
	expect, _ := hex.DecodeString("02000000000000000c")
	require.Equal(t, expect, ReportStoreKey(12))
}

func TestDataSourceStoreKey(t *testing.T) {
	expect, _ := hex.DecodeString("030000000000000378")
	require.Equal(t, expect, DataSourceStoreKey(888))
}

func TestOracleScriptStoreKey(t *testing.T) {
	expect, _ := hex.DecodeString("04000000000000007b")
	require.Equal(t, expect, OracleScriptStoreKey(123))
}

func TestValidatorStatusStoreKey(t *testing.T) {
	val, _ := sdk.ValAddressFromHex("b80f2a5df7d5710b15622d1a9f1e3830ded5bda8")
	expect, _ := hex.DecodeString("05b80f2a5df7d5710b15622d1a9f1e3830ded5bda8")
	require.Equal(t, expect, ValidatorStatusStoreKey(val))
}

func TestResultStoreKey(t *testing.T) {
	expect, _ := hex.DecodeString("ff0000000000000014")
	require.Equal(t, expect, ResultStoreKey(20))
}

func TestReportsOfValidatorPrefixKey(t *testing.T) {
	val, _ := sdk.ValAddressFromHex("b80f2a5df7d5710b15622d1a9f1e3830ded5bda8")
	expect, _ := hex.DecodeString("020000000000000014b80f2a5df7d5710b15622d1a9f1e3830ded5bda8")
	require.Equal(t, expect, ReportsOfValidatorPrefixKey(20, val))
}
