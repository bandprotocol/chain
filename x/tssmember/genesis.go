package tssmember

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tssmember/keeper"
	"github.com/bandprotocol/chain/v2/x/tssmember/types"
)

// InitGenesis performs genesis initialization for the tss module.
func InitGenesis(ctx sdk.Context, k *keeper.Keeper, data *types.GenesisState) {
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
	}
}
