package tss

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tss/keeper"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// InitGenesis performs genesis initialization for the tss module.
func InitGenesis(ctx sdk.Context, k *keeper.Keeper, data *types.GenesisState) {
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	k.SetGroupCount(ctx, data.GroupCount)
	for _, group := range data.Groups {
		k.SetGroup(ctx, group)
	}

	k.SetMembers(ctx, data.Members)

	k.SetSigningCount(ctx, data.SigningCount)
	for _, signing := range data.Signings {
		k.SetSigning(ctx, signing)
	}

	desMapping := make(map[string][]types.DE)
	for _, deq := range data.DEsGenesis {
		desMapping[deq.Address] = append(desMapping[deq.Address], deq.DE)
	}

	for addr, des := range desMapping {
		if uint64(len(des)) > data.Params.MaxDESize {
			panic(fmt.Sprintf("DEsGenesis of %s size exceeds MaxDESize", addr))
		}

		acc := sdk.MustAccAddressFromBech32(addr)
		k.SetDECount(ctx, acc, uint64(len(des)))

		for _, de := range des {
			k.SetDE(ctx, acc, de)
		}
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:       k.GetParams(ctx),
		GroupCount:   k.GetGroupCount(ctx),
		Groups:       k.GetGroups(ctx),
		Members:      k.GetMembers(ctx),
		SigningCount: k.GetSigningCount(ctx),
		Signings:     k.GetSignings(ctx),
		DEsGenesis:   k.GetDEsGenesis(ctx),
	}
}
