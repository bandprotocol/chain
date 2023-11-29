package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// GenMaxRawRequestCount returns randomize MaxRawRequestCount
func GenMaxRawRequestCount(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 1, 100))
}

// GenMaxAskCount returns randomize MaxAskCount
func GenMaxAskCount(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 10, 50))
}

// GenMaxCalldataSize returns randomize MaxCalldataSize
func GenMaxCalldataSize(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 100, 1000))
}

// GenMaxReportDataSize returns randomize MaxReportDataSize
func GenMaxReportDataSize(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 100, 1000))
}

// GenExpirationBlockCount returns randomize ExpirationBlockCount
func GenExpirationBlockCount(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 10, 1000))
}

// GenBaseOwasmGas returns randomize BaseOwasmGas
func GenBaseOwasmGas(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 0, 300000))
}

// GenPerValidatorRequestGas returns randomize PerValidatorRequestGas
func GenPerValidatorRequestGas(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 0, 10000))
}

// GenSamplingTryCount returns randomize SamplingTryCount
func GenSamplingTryCount(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 1, 10))
}

// GenOracleRewardPercentage returns randomize OracleRewardPercentage
func GenOracleRewardPercentage(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 0, 100))
}

// GenInactivePenaltyDuration returns randomize InactivePenaltyDuration
func GenInactivePenaltyDuration(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 10000000000, 1000000000000))
}

// GenIBCRequestEnabled returns randomized IBCRequestEnabled
func GenIBCRequestEnabled(r *rand.Rand) bool {
	return r.Int63n(100) < 50
}

// RandomizedGenState generates a random GenesisState for oracle
func RandomizedGenState(simState *module.SimulationState) {
	var maxRawRequestCount uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, string(types.KeyMaxRawRequestCount), &maxRawRequestCount, simState.Rand,
		func(r *rand.Rand) { maxRawRequestCount = GenMaxRawRequestCount(r) },
	)

	var maxAskCount uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, string(types.KeyMaxAskCount), &maxAskCount, simState.Rand,
		func(r *rand.Rand) { maxAskCount = GenMaxAskCount(r) },
	)

	var maxCalldataSize uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, string(types.KeyMaxCalldataSize), &maxCalldataSize, simState.Rand,
		func(r *rand.Rand) { maxCalldataSize = GenMaxCalldataSize(r) },
	)

	var maxReportDataSize uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, string(types.KeyMaxReportDataSize), &maxReportDataSize, simState.Rand,
		func(r *rand.Rand) { maxReportDataSize = GenMaxReportDataSize(r) },
	)

	var expirationBlockCount uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, string(types.KeyExpirationBlockCount), &expirationBlockCount, simState.Rand,
		func(r *rand.Rand) { expirationBlockCount = GenExpirationBlockCount(r) },
	)

	var baseOwasmGas uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, string(types.KeyBaseOwasmGas), &baseOwasmGas, simState.Rand,
		func(r *rand.Rand) { baseOwasmGas = GenBaseOwasmGas(r) },
	)

	var perValidatorRequestGas uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, string(types.KeyPerValidatorRequestGas), &perValidatorRequestGas, simState.Rand,
		func(r *rand.Rand) { perValidatorRequestGas = GenPerValidatorRequestGas(r) },
	)

	var samplingTryCount uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, string(types.KeySamplingTryCount), &samplingTryCount, simState.Rand,
		func(r *rand.Rand) { samplingTryCount = GenSamplingTryCount(r) },
	)

	var oracleRewardPercentage uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, string(types.KeyOracleRewardPercentage), &oracleRewardPercentage, simState.Rand,
		func(r *rand.Rand) { oracleRewardPercentage = GenOracleRewardPercentage(r) },
	)

	var inactivePenaltyDuration uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, string(types.KeyInactivePenaltyDuration), &inactivePenaltyDuration, simState.Rand,
		func(r *rand.Rand) { inactivePenaltyDuration = GenInactivePenaltyDuration(r) },
	)

	var ibcRequestEnabled bool
	simState.AppParams.GetOrGenerate(
		simState.Cdc, string(types.KeyIBCRequestEnabled), &ibcRequestEnabled, simState.Rand,
		func(r *rand.Rand) { ibcRequestEnabled = GenIBCRequestEnabled(r) },
	)

	oracleGenesis := types.NewGenesisState(
		types.NewParams(
			maxRawRequestCount,
			maxAskCount,
			maxCalldataSize,
			maxReportDataSize,
			expirationBlockCount,
			baseOwasmGas,
			perValidatorRequestGas,
			samplingTryCount,
			oracleRewardPercentage,
			inactivePenaltyDuration,
			ibcRequestEnabled,
		),
		[]types.DataSource{},
		[]types.OracleScript{},
	)

	bz, err := json.MarshalIndent(&oracleGenesis, "", " ")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Selected randomly generated oracle parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(oracleGenesis)
}
