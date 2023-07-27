package tss

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tss/keeper"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// InitGenesis performs genesis initialization for the tss module.
func InitGenesis(ctx sdk.Context, k *keeper.Keeper, data *types.GenesisState) {
	k.SetParams(ctx, data.Params)
	k.SetGroupCount(ctx, data.GroupCount)
	k.SetSigningCount(ctx, data.SigningCount)
	for _, group := range data.Groups {
		k.SetGroup(ctx, group)
	}
	for _, deq := range data.DEQueuesGenesis {
		k.SetDEQueue(ctx, deq.Address, *deq.DEQueue)
	}
	for _, de := range data.DEsGenesis {
		k.SetDE(ctx, de.Address, de.Index, *de.DE)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:          k.GetParams(ctx),
		GroupCount:      k.GetGroupCount(ctx),
		SigningCount:    k.GetSigningCount(ctx),
		Groups:          k.GetGroups(ctx),
		DEQueuesGenesis: k.GetDEQueuesGenesis(ctx),
		DEsGenesis:      k.GetDEsGenesis(ctx),
	}
}
