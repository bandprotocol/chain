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

	for _, status := range data.Statuses {
		k.SetStatus(ctx, status)
	}

	for _, signingFee := range data.SigningFees {
		k.SetSigningFee(ctx, signingFee)
	}

	k.SetCurrentGroupID(ctx, data.CurrentGroupID)
	k.SetReplacingGroupID(ctx, data.ReplacingGroupID)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:           k.GetParams(ctx),
		Statuses:         k.GetStatuses(ctx),
		CurrentGroupID:   k.GetCurrentGroupID(ctx),
		ReplacingGroupID: k.GetReplacingGroupID(ctx),
		SigningFees:      k.GetSigningFees(ctx),
	}
}
