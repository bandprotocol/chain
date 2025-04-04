package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/bandprotocol/chain/v3/x/oracle"
	"github.com/bandprotocol/chain/v3/x/oracle/simulation"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

func TestRandomizedGenState(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(oracle.AppModuleBasic{}).Codec
	s := rand.NewSource(1)
	r := rand.New(s)

	simState := module.SimulationState{
		AppParams:    make(simtypes.AppParams),
		Cdc:          cdc,
		Rand:         r,
		NumBonded:    3,
		Accounts:     simtypes.RandomAccounts(r, 3),
		InitialStake: sdkmath.NewInt(1000),
		GenState:     make(map[string]json.RawMessage),
	}

	simulation.RandomizedGenState(&simState)

	var oracleGenesis types.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[types.ModuleName], &oracleGenesis)

	require.Equal(t, uint64(18), oracleGenesis.Params.MaxRawRequestCount)
	require.Equal(t, uint64(26), oracleGenesis.Params.MaxAskCount)
	require.Equal(t, uint64(700), oracleGenesis.Params.MaxCalldataSize)
	require.Equal(t, uint64(294), oracleGenesis.Params.MaxReportDataSize)
	require.Equal(t, uint64(791), oracleGenesis.Params.ExpirationBlockCount)
	require.Equal(t, uint64(228162), oracleGenesis.Params.BaseOwasmGas)
	require.Equal(t, uint64(5089), oracleGenesis.Params.PerValidatorRequestGas)
	require.Equal(t, uint64(5), oracleGenesis.Params.SamplingTryCount)
	require.Equal(t, uint64(74), oracleGenesis.Params.OracleRewardPercentage)
	require.Equal(t, uint64(265472644968), oracleGenesis.Params.InactivePenaltyDuration)
	require.Equal(t, false, oracleGenesis.Params.IBCRequestEnabled)
	require.Equal(t, []types.DataSource{}, oracleGenesis.DataSources)
	require.Equal(t, []types.OracleScript{}, oracleGenesis.OracleScripts)
}

// TestRandomizedGenState1 tests abnormal scenarios of applying RandomizedGenState.
func TestRandomizedGenState1(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	s := rand.NewSource(1)
	r := rand.New(s)
	// all these tests will panic
	tests := []struct {
		simState module.SimulationState
		panicMsg string
	}{
		{ // panic => reason: incomplete initialization of the simState
			module.SimulationState{}, "invalid memory address or nil pointer dereference"},
		{ // panic => reason: incomplete initialization of the simState
			module.SimulationState{
				AppParams: make(simtypes.AppParams),
				Cdc:       cdc,
				Rand:      r,
			}, "assignment to entry in nil map"},
	}

	for _, tt := range tests {
		temp := tt
		require.Panicsf(t, func() { simulation.RandomizedGenState(&temp.simState) }, tt.panicMsg)
	}
}
