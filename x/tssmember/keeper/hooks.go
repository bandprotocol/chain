package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// Wrapper struct
type Hooks struct {
	k Keeper
}

var _ tsstypes.TSSHooks = Hooks{}

// Create new distribution hooks
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) AfterGroupActivated(ctx sdk.Context, group tsstypes.Group) {}

func (h Hooks) AfterGroupFailedToActivate(ctx sdk.Context, group tsstypes.Group) {}

func (h Hooks) AfterGroupReplaced(ctx sdk.Context, replacement tsstypes.Replacement) {}

func (h Hooks) AfterGroupFailedToReplace(ctx sdk.Context, replacement tsstypes.Replacement) {}

func (h Hooks) AfterStatusUpdated(ctx sdk.Context, status tsstypes.Status) {}
