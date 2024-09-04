package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

func TestGetSetRequestCount(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, true)
	k := app.OracleKeeper

	// Initially request count must be 0.
	require.Equal(t, uint64(0), k.GetRequestCount(ctx))
	// After we set the count manually, it should be reflected.
	k.SetRequestCount(ctx, 42)
	require.Equal(t, uint64(42), k.GetRequestCount(ctx))
}

func TestGetDataSourceCount(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, true)
	k := app.OracleKeeper

	k.SetDataSourceCount(ctx, 42)
	require.Equal(t, uint64(42), k.GetDataSourceCount(ctx))
}

func TestGetSetOracleScriptCount(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, true)
	k := app.OracleKeeper

	k.SetOracleScriptCount(ctx, 42)
	require.Equal(t, uint64(42), k.GetOracleScriptCount(ctx))
}

func TestGetSetRollingSeed(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, true)
	k := app.OracleKeeper

	k.SetRollingSeed(ctx, []byte("HELLO_WORLD"))
	require.Equal(t, []byte("HELLO_WORLD"), k.GetRollingSeed(ctx))
}

func TestGetNextRequestID(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, true)
	k := app.OracleKeeper

	// First request id must be 1.
	require.Equal(t, types.RequestID(1), k.GetNextRequestID(ctx))
	// After we add new requests, the request count must increase accordingly.
	require.Equal(t, uint64(1), k.GetRequestCount(ctx))
	require.Equal(t, types.RequestID(2), k.GetNextRequestID(ctx))
	require.Equal(t, types.RequestID(3), k.GetNextRequestID(ctx))
	require.Equal(t, types.RequestID(4), k.GetNextRequestID(ctx))
	require.Equal(t, uint64(4), k.GetRequestCount(ctx))
}

func TestGetNextDataSourceID(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, true)
	k := app.OracleKeeper

	initialID := k.GetDataSourceCount(ctx)
	require.Equal(t, types.DataSourceID(initialID+1), k.GetNextDataSourceID(ctx))
	require.Equal(t, types.DataSourceID(initialID+2), k.GetNextDataSourceID(ctx))
	require.Equal(t, types.DataSourceID(initialID+3), k.GetNextDataSourceID(ctx))
}

func TestGetNextOracleScriptID(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, true)
	k := app.OracleKeeper

	initialID := k.GetOracleScriptCount(ctx)
	require.Equal(t, types.OracleScriptID(initialID+1), k.GetNextOracleScriptID(ctx))
	require.Equal(t, types.OracleScriptID(initialID+2), k.GetNextOracleScriptID(ctx))
	require.Equal(t, types.OracleScriptID(initialID+3), k.GetNextOracleScriptID(ctx))
}

func TestGetSetRequestLastExpiredID(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, true)
	k := app.OracleKeeper

	// Initially last expired request must be 0.
	require.Equal(t, types.RequestID(0), k.GetRequestLastExpired(ctx))
	k.SetRequestLastExpired(ctx, 20)
	require.Equal(t, types.RequestID(20), k.GetRequestLastExpired(ctx))
}
