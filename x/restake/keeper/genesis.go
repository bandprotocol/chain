package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	for _, key := range data.Keys {
		k.SetKey(ctx, key)
	}

	for _, stake := range data.Stakes {
		k.SetStake(ctx, stake)
	}
}

// ExportGenesis returns the module's exported genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Keys:   k.GetKeys(ctx),
		Stakes: k.GetStakes(ctx),
	}
}
