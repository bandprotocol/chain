package auction

import (
	auctionkeeper "github.com/GeoDB-Limited/odin-core/x/auction/keeper"
	auctiontypes "github.com/GeoDB-Limited/odin-core/x/auction/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis new mint genesis
func InitGenesis(ctx sdk.Context, keeper auctionkeeper.Keeper, data *auctiontypes.GenesisState) {
	keeper.SetParams(ctx, data.Params)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper auctionkeeper.Keeper) *auctiontypes.GenesisState {
	params := keeper.GetParams(ctx)
	return auctiontypes.NewGenesisState(params)
}
