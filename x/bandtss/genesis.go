package bandtss

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/bandtss/keeper"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

// InitGenesis performs genesis initialization for this module.
func InitGenesis(ctx sdk.Context, k *keeper.Keeper, data *types.GenesisState) {
	// check if the module account exists
	moduleAcc := k.GetBandtssAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// Set module account if its balance is zero
	if balance := k.GetModuleBalance(ctx); balance.IsZero() {
		k.SetModuleAccount(ctx, moduleAcc)
	}

	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	for _, member := range data.Members {
		k.SetMember(ctx, member)
	}

	k.SetSigningCount(ctx, data.SigningCount)
	for _, signingInfo := range data.Signings {
		k.SetSigning(ctx, signingInfo)
	}

	for _, mapping := range data.SigningIDMappings {
		k.SetSigningIDMapping(ctx, mapping.SigningID, mapping.BandtssSigningID)
	}

	k.SetCurrentGroupID(ctx, data.CurrentGroupID)
	k.SetReplacement(ctx, data.Replacement)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:            k.GetParams(ctx),
		Members:           k.GetMembers(ctx),
		CurrentGroupID:    k.GetCurrentGroupID(ctx),
		SigningCount:      k.GetSigningCount(ctx),
		Signings:          k.GetSignings(ctx),
		SigningIDMappings: k.GetSigningIDMappings(ctx),
		Replacement:       k.GetReplacement(ctx),
	}
}
