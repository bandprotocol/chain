package rollingseed

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/rollingseed/keeper"
	"github.com/bandprotocol/chain/v2/x/rollingseed/types"
)

// InitGenesis performs genesis initialization for the rollingseed module.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	k.SetRollingSeed(ctx, make([]byte, types.RollingSeedSizeInBytes))
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{}
}
