package keeper

import (
	"fmt"
	auctiontypes "github.com/GeoDB-Limited/odin-core/x/auction/types"
	minttypes "github.com/GeoDB-Limited/odin-core/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis new mint genesis
func InitGenesis(ctx sdk.Context, keeper Keeper, data *auctiontypes.GenesisState) {
	keeper.SetParams(ctx, data.Params)

	moduleAcc := keeper.GetAuctionAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", minttypes.ModuleName))
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) *auctiontypes.GenesisState {
	params := keeper.GetParams(ctx)
	return auctiontypes.NewGenesisState(params)
}
