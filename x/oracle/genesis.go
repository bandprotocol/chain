package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/bandprotocol/chain/x/oracle/keeper"
	"github.com/bandprotocol/chain/x/oracle/types"
)

// InitGenesis performs genesis initialization for the oracle module.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) []abci.ValidatorUpdate {
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
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:        k.GetParams(ctx),
		DataSources:   k.GetAllDataSources(ctx),
		OracleScripts: k.GetAllOracleScripts(ctx),
	}
}
