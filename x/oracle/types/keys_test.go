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

func TestReporterStoreKey(t *testing.T) {
	val, _ := sdk.ValAddressFromHex("b80f2a5df7d5710b15622d1a9f1e3830ded5bda8")
	rep, _ := sdk.AccAddressFromHex("ba11d00c5f74255f56a5e366f4f77f5a186d7f55")
	expect, _ := hex.DecodeString("05b80f2a5df7d5710b15622d1a9f1e3830ded5bda8ba11d00c5f74255f56a5e366f4f77f5a186d7f55")
	require.Equal(t, expect, ReporterStoreKey(val, rep))
}

func TestValidatorStatusStoreKey(t *testing.T) {
	val, _ := sdk.ValAddressFromHex("b80f2a5df7d5710b15622d1a9f1e3830ded5bda8")
	expect, _ := hex.DecodeString("06b80f2a5df7d5710b15622d1a9f1e3830ded5bda8")
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

func TestReportersOfValidatorPrefixKey(t *testing.T) {
	val, _ := sdk.ValAddressFromHex("b80f2a5df7d5710b15622d1a9f1e3830ded5bda8")
	expect, _ := hex.DecodeString("05b80f2a5df7d5710b15622d1a9f1e3830ded5bda8")
	require.Equal(t, expect, ReportersOfValidatorPrefixKey(val))
}

func TestGetEscrowAddress(t *testing.T) {
	var (
		requestKey1 = "beeb"
		port1       = "transfer"
		channel1    = "channel"
		requestKey2 = "beeb"
		port2       = "transfercha"
		channel2    = "nnel"
	)

	escrow1 := GetEscrowAddress(requestKey1, port1, channel1)
	escrow2 := GetEscrowAddress(requestKey2, port2, channel2)
	require.NotEqual(t, escrow1, escrow2)
}
