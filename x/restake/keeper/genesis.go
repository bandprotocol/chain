package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/restake/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	for _, vault := range data.Vaults {
		k.SetVault(ctx, vault)
	}

	for _, lock := range data.Locks {
		k.SetLock(ctx, lock)
	}
}

// ExportGenesis returns the module's exported genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return types.NewGenesisState(
		k.GetVaults(ctx),
		k.GetLocks(ctx),
	)
}
