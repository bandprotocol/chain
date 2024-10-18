package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	k.SetAllDelegatorSignals(ctx, genState.DelegatorSignals)

	signalTotalPowers := k.CalculateNewSignalTotalPowers(ctx)
	k.SetSignalTotalPowers(ctx, signalTotalPowers)

	feeds := k.CalculateNewCurrentFeeds(ctx)
	k.SetCurrentFeeds(ctx, feeds)

	if err := k.SetReferenceSourceConfig(ctx, genState.ReferenceSourceConfig); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the module's exported genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return types.NewGenesisState(k.GetParams(ctx), k.GetAllDelegatorSignals(ctx), k.GetReferenceSourceConfig(ctx))
}
