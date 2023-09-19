package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/globalfee/types"
)

// InitGenesis new globalfee genesis
func (keeper Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	if err := keeper.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}
}

// ExportGenesis returns a GenesisState for a given context.
func (keeper Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := keeper.GetParams(ctx)
	return types.NewGenesisState(params)
}
