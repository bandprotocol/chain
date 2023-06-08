package tss

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tss/keeper"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// InitGenesis performs genesis initialization for the tss module.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	k.SetParams(ctx, data.Params)
	k.SetGroupCount(ctx, 0)
	k.SetSigningCount(ctx, 0)
	k.SetRollingSeed(ctx, make([]byte, types.RollingSeedSizeInBytes))
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
	}
}
