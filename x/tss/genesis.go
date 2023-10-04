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
	for _, group := range data.Groups {
		k.SetGroup(ctx, group)
	}

	for _, member := range data.Members {
		k.SetMember(ctx, member)
	}

	k.SetSigningCount(ctx, data.SigningCount)
	for _, signing := range data.Signings {
		k.SetSigning(ctx, signing)
	}

	k.SetReplacementCount(ctx, data.ReplacementCount)
	for _, rep := range data.Replacements {
		k.SetReplacement(ctx, rep)
	}

	for _, deq := range data.DEQueuesGenesis {
		address := sdk.MustAccAddressFromBech32(deq.Address)
		k.SetDEQueue(ctx, address, deq.DEQueue)
	}

	for _, de := range data.DEsGenesis {
		address := sdk.MustAccAddressFromBech32(de.Address)
		k.SetDE(ctx, address, de.Index, de.DE)
	}

	for _, status := range data.Statuses {
		k.SetMemberStatus(ctx, status)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:           k.GetParams(ctx),
		GroupCount:       k.GetGroupCount(ctx),
		Groups:           k.GetGroups(ctx),
		Members:          k.GetMembers(ctx),
		SigningCount:     k.GetSigningCount(ctx),
		Signings:         k.GetAllReplacementSigning(ctx),
		ReplacementCount: k.GetReplacementCount(ctx),
		Replacements:     k.GetReplacements(ctx),
		DEQueuesGenesis:  k.GetDEQueuesGenesis(ctx),
		DEsGenesis:       k.GetDEsGenesis(ctx),
		Statuses:         k.GetStatuses(ctx),
	}
}
