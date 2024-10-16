package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/rollingseed/types"
)

// InitGenesis performs genesis initialization for the rollingseed module.
func (k *Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	k.SetRollingSeed(ctx, make([]byte, types.RollingSeedSizeInBytes))
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (k *Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{}
}
