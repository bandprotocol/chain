package oracle

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/oracle/keeper"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// InitGenesis performs genesis initialization for the oracle module.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	k.SetParams(ctx, data.Params)
	k.SetDataSourceCount(ctx, 0)
	k.SetOracleScriptCount(ctx, 0)
	k.SetRequestCount(ctx, 0)
	k.SetRequestLastExpired(ctx, 0)
	k.SetRollingSeed(ctx, make([]byte, types.RollingSeedSizeInBytes))
	for _, dataSource := range data.DataSources {
		_ = k.AddDataSource(ctx, dataSource)
	}
	for _, oracleScript := range data.OracleScripts {
		_ = k.AddOracleScript(ctx, oracleScript)
	}
	for _, reportersPerValidator := range data.Reporters {
		valAddr, err := sdk.ValAddressFromBech32(reportersPerValidator.Validator)
		if err != nil {
			panic(fmt.Sprintf("unable to parse validator address %s: %v", reportersPerValidator.Validator, err))
		}
		for _, reporterBech32 := range reportersPerValidator.Reporters {
			reporterAddr, err := sdk.AccAddressFromBech32(reporterBech32)
			if err != nil {
				panic(fmt.Sprintf("unable to parse reporter address %s: %v", reporterBech32, err))
			}
			if valAddr.Equals(sdk.ValAddress(reporterAddr)) {
				continue
			}
			err = k.AddReporter(ctx, valAddr, reporterAddr)
			if err != nil {
				panic(fmt.Sprintf("unable to add reporter %s: %v", reporterBech32, err))
			}
		}
	}

	k.SetPort(ctx, types.PortID)
	// Only try to bind to port if it is not already bound, since we may already own
	// port capability from capability InitGenesis
	if !k.IsBound(ctx, types.PortID) {
		// transfer module binds to the transfer port on InitChain
		// and claims the returned capability
		err := k.BindPort(ctx, types.PortID)
		if err != nil {
			panic(fmt.Sprintf("could not claim port capability: %v", err))
		}
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:        k.GetParams(ctx),
		DataSources:   k.GetAllDataSources(ctx),
		OracleScripts: k.GetAllOracleScripts(ctx),
		Reporters:     k.GetAllReporters(ctx),
	}
}
