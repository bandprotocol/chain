package simulation_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/bandprotocol/chain/v3/x/oracle"
	"github.com/bandprotocol/chain/v3/x/oracle/simulation"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

var (
	accAddr  = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	valAddr1 = sdk.ValAddress(ed25519.GenPrivKey().PubKey().Address())
	valAddr2 = sdk.ValAddress(ed25519.GenPrivKey().PubKey().Address())
	valAddr3 = sdk.ValAddress(ed25519.GenPrivKey().PubKey().Address())
	treaAddr = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
)

func TestDecodeStore(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(oracle.AppModuleBasic{}).Codec
	dec := simulation.NewDecodeStore(cdc)

	rawRequest := types.NewRawRequest(1, 1, []byte("calldata"))
	request := types.NewRequest(
		1,
		[]byte("calldata"),
		[]sdk.ValAddress{valAddr1, valAddr2, valAddr3},
		2,
		1,
		time.Now().UTC(),
		"client",
		[]types.RawRequest{rawRequest},
		nil,
		100000,
	)

	rawReport := types.NewRawReport(1, 0, []byte("data"))
	report := types.NewReport(valAddr1, true, []types.RawReport{rawReport})

	dataSource := types.NewDataSource(
		accAddr,
		"name",
		"description",
		"filename",
		sdk.NewCoins(sdk.NewInt64Coin("band", 1000)),
		treaAddr,
	)

	oracleScript := types.NewOracleScript(
		accAddr,
		"name",
		"description",
		"filename",
		"{symbols:[string],multiplier:u64}/{rates:[u64]}",
		"https://url.com",
	)

	status := types.NewValidatorStatus(true, time.Now().UTC())

	result := types.NewResult(
		"client",
		1,
		[]byte("calldata"),
		3,
		2,
		1,
		1,
		1000,
		1000,
		types.RESOLVE_STATUS_SUCCESS,
		[]byte("result"),
	)

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.RequestStoreKey(1), Value: cdc.MustMarshal(&request)},
			{Key: types.ReportStoreKey(1), Value: cdc.MustMarshal(&report)},
			{Key: types.DataSourceStoreKey(1), Value: cdc.MustMarshal(&dataSource)},
			{Key: types.OracleScriptStoreKey(1), Value: cdc.MustMarshal(&oracleScript)},
			{Key: types.ValidatorStatusStoreKey(valAddr1), Value: cdc.MustMarshal(&status)},
			{Key: types.ResultStoreKey(1), Value: cdc.MustMarshal(&result)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Request", fmt.Sprintf("%v\n%v", request, request)},
		{"Report", fmt.Sprintf("%v\n%v", report, report)},
		{"DataSource", fmt.Sprintf("%v\n%v", dataSource, dataSource)},
		{"OracleScript", fmt.Sprintf("%v\n%v", oracleScript, oracleScript)},
		{"ValidatorStatus", fmt.Sprintf("%v\n%v", status, status)},
		{"Result", fmt.Sprintf("%v\n%v", result, result)},
		{"other", ""},
	}

	for idx, test := range tests {
		i, tt := idx, test
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
