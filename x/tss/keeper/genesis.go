package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// InitGenesis performs genesis initialization for the tss module.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	k.SetGroupGenesis(ctx, data.Groups)

	k.SetMembers(ctx, data.Members)

	desMapping := make(map[string][]types.DE)
	for _, deq := range data.DEs {
		desMapping[deq.Address] = append(desMapping[deq.Address], deq.DE)
	}

	for addr, des := range desMapping {
		if uint64(len(des)) > data.Params.MaxDESize {
			panic(fmt.Sprintf("DEs of %s size exceeds MaxDESize", addr))
		}

		acc := sdk.MustAccAddressFromBech32(addr)
		deQueue := types.NewDEQueue(0, uint64(len(des)))
		k.SetDEQueue(ctx, acc, deQueue)

		for i, de := range des {
			k.SetDE(ctx, acc, uint64(i), de)
		}
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:  k.GetParams(ctx),
		Groups:  k.GetGroups(ctx),
		Members: k.GetMembers(ctx),
		DEs:     k.GetDEsGenesis(ctx),
	}
}
