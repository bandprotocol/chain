package bandtss

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/bandtss/keeper"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

// InitGenesis performs genesis initialization for the bandtss module.
func InitGenesis(ctx sdk.Context, k *keeper.Keeper, data *types.GenesisState) {
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	for _, member := range data.Members {
		k.SetMember(ctx, member)
	}

	for _, signingInfo := range data.Signings {
		k.SetSigning(ctx, signingInfo)
	}

	for _, mapping := range data.SigningIDMappings {
		k.SetSigningIDMapping(ctx, mapping.SigningID, mapping.BandtssSigningID)
	}

	k.SetCurrentGroupID(ctx, data.CurrentGroupID)
	k.SetReplacingGroupID(ctx, data.ReplacingGroupID)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:            k.GetParams(ctx),
		Members:           k.GetMembers(ctx),
		CurrentGroupID:    k.GetCurrentGroupID(ctx),
		ReplacingGroupID:  k.GetReplacingGroupID(ctx),
		Signings:          k.GetSignings(ctx),
		SigningIDMappings: k.GetSigningIDMappings(ctx),
	}
}
