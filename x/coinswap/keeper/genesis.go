package keeper

import (
	"encoding/json"
	coinswaptypes "github.com/GeoDB-Limited/odin-core/x/coinswap/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// DefaultGenesisState returns the default oracle genesis state.
func DefaultGenesisState() *coinswaptypes.GenesisState {
	return &coinswaptypes.GenesisState{
		Params: coinswaptypes.DefaultParams(),
	}
}

// InitGenesis performs genesis initialization for the oracle module.
func InitGenesis(ctx sdk.Context, k Keeper, data coinswaptypes.GenesisState) []abci.ValidatorUpdate {
	k.SetParams(ctx, data.Params)
	k.SetInitialRate(ctx, data.InitialRate)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k Keeper) coinswaptypes.GenesisState {
	return coinswaptypes.GenesisState{
		Params: k.GetParams(ctx),
	}
}

// GetGenesisStateFromAppState returns x/coinswap GenesisState given raw application genesis state.
func GetGenesisStateFromAppState(cdc *codec.LegacyAmino, appState map[string]json.RawMessage) coinswaptypes.GenesisState {
	var genesisState coinswaptypes.GenesisState
	if appState[coinswaptypes.ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[coinswaptypes.ModuleName], &genesisState)
	}
	return genesisState
}
