package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/x/oracle/testapp"
	"github.com/bandprotocol/chain/x/oracle/types"
)

func TestGetSetRequestCount(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	// Initially request count must be 0.
	require.Equal(t, int64(0), k.GetRequestCount(ctx))
	// After we set the count manually, it should be reflected.
	k.SetRequestCount(ctx, 42)
	require.Equal(t, int64(42), k.GetRequestCount(ctx))
}

func TestGetDataSourceCount(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	k.SetDataSourceCount(ctx, 42)
	require.Equal(t, int64(42), k.GetDataSourceCount(ctx))
}

func TestGetSetOracleScriptCount(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	k.SetOracleScriptCount(ctx, 42)
	require.Equal(t, int64(42), k.GetOracleScriptCount(ctx))
}

func TestGetSetRollingSeed(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	k.SetRollingSeed(ctx, []byte("HELLO_WORLD"))
	require.Equal(t, []byte("HELLO_WORLD"), k.GetRollingSeed(ctx))
}

func TestGetNextRequestID(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	// First request id must be 1.
	require.Equal(t, types.RequestID(1), k.GetNextRequestID(ctx))
	// After we add new requests, the request count must increase accordingly.
	require.Equal(t, int64(1), k.GetRequestCount(ctx))
	require.Equal(t, types.RequestID(2), k.GetNextRequestID(ctx))
	require.Equal(t, types.RequestID(3), k.GetNextRequestID(ctx))
	require.Equal(t, types.RequestID(4), k.GetNextRequestID(ctx))
	require.Equal(t, int64(4), k.GetRequestCount(ctx))
}

func TestGetNextDataSourceID(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	initialID := k.GetDataSourceCount(ctx)
	require.Equal(t, types.DataSourceID(initialID+1), k.GetNextDataSourceID(ctx))
	require.Equal(t, types.DataSourceID(initialID+2), k.GetNextDataSourceID(ctx))
	require.Equal(t, types.DataSourceID(initialID+3), k.GetNextDataSourceID(ctx))
}

func TestGetNextOracleScriptID(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	initialID := k.GetOracleScriptCount(ctx)
	require.Equal(t, types.OracleScriptID(initialID+1), k.GetNextOracleScriptID(ctx))
	require.Equal(t, types.OracleScriptID(initialID+2), k.GetNextOracleScriptID(ctx))
	require.Equal(t, types.OracleScriptID(initialID+3), k.GetNextOracleScriptID(ctx))
}

func TestGetSetRequestLastExpiredID(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	// Initially last expired request must be 0.
	require.Equal(t, types.RequestID(0), k.GetRequestLastExpired(ctx))
	k.SetRequestLastExpired(ctx, 20)
	require.Equal(t, types.RequestID(20), k.GetRequestLastExpired(ctx))
}

func TestGetSetParams(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	k.SetUInt64Param(ctx, types.KeyMaxRawRequestCount, 1)
	k.SetUInt64Param(ctx, types.KeyMaxAskCount, 10)
	k.SetUInt64Param(ctx, types.KeyExpirationBlockCount, 30)
	k.SetUInt64Param(ctx, types.KeyBaseOwasmGas, 50000)
	k.SetUInt64Param(ctx, types.KeyPerValidatorRequestGas, 3000)
	k.SetUInt64Param(ctx, types.KeySamplingTryCount, 3)
	k.SetUInt64Param(ctx, types.KeyOracleRewardPercentage, 50)
	k.SetUInt64Param(ctx, types.KeyInactivePenaltyDuration, 1000)
	k.SetBoolParam(ctx, types.KeyIBCRequestEnabled, true)
	require.Equal(t, types.NewParams(1, 10, 30, 50000, 3000, 3, 50, 1000, true), k.GetParams(ctx))
	k.SetUInt64Param(ctx, types.KeyMaxRawRequestCount, 2)
	k.SetUInt64Param(ctx, types.KeyMaxAskCount, 20)
	k.SetUInt64Param(ctx, types.KeyExpirationBlockCount, 40)
	k.SetUInt64Param(ctx, types.KeyBaseOwasmGas, 150000)
	k.SetUInt64Param(ctx, types.KeyPerValidatorRequestGas, 30000)
	k.SetUInt64Param(ctx, types.KeySamplingTryCount, 5)
	k.SetUInt64Param(ctx, types.KeyOracleRewardPercentage, 80)
	k.SetUInt64Param(ctx, types.KeyInactivePenaltyDuration, 10000)
	k.SetBoolParam(ctx, types.KeyIBCRequestEnabled, false)
	require.Equal(t, types.NewParams(2, 20, 40, 150000, 30000, 5, 80, 10000, false), k.GetParams(ctx))
}
