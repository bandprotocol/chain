package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
)

// InitGenesis performs genesis initialization for this module.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
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

	if data.CurrentGroup.GroupID != tss.GroupID(0) &&
		data.CurrentGroup.ActiveTime.After(ctx.BlockTime()) {
		panic("invalid genesis state: current group active time is in the future")
	}

	if data.CurrentGroup.GroupID == tss.GroupID(0) &&
		!data.CurrentGroup.ActiveTime.IsZero() {
		panic("invalid genesis state: current group ID is not set but active time is set")
	}

	k.SetCurrentGroup(ctx, data.CurrentGroup)

	for _, member := range data.Members {
		k.SetMember(ctx, member)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:       k.GetParams(ctx),
		Members:      k.GetMembers(ctx),
		CurrentGroup: k.GetCurrentGroup(ctx),
	}
}
