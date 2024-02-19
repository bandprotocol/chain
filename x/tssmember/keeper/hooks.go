package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
	"github.com/bandprotocol/chain/v2/x/tssmember/types"
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

func (h Hooks) AfterSigningFailed(ctx sdk.Context, signing tsstypes.Signing) {
	if signing.Fee.IsZero() {
		return
	}

	address := sdk.MustAccAddressFromBech32(signing.Requester)
	feeCoins := signing.Fee.MulInt(sdk.NewInt(int64(len(signing.AssignedMembers))))

	// Refund fee to requester
	err := h.k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, feeCoins)
	if err != nil {
		panic(err) // Error is not possible
	}
}

func (h Hooks) AfterSigningCompleted(ctx sdk.Context, signing tsstypes.Signing) {
	// Send fee to assigned members.
	for _, am := range signing.AssignedMembers {
		address := sdk.MustAccAddressFromBech32(am.Address)
		if err := h.k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, signing.Fee); err != nil {
			panic(err) // Error is not possible
		}
	}
}

func (h Hooks) AfterSigningInitiated(ctx sdk.Context, signing tsstypes.Signing) error {
	feeCoins := signing.Fee.MulInt(sdk.NewInt(int64(len(signing.AssignedMembers))))
	if feeCoins.IsZero() {
		return nil
	}

	address, err := sdk.AccAddressFromBech32(signing.Requester)
	if err != nil {
		return err
	}

	err = h.k.bankKeeper.SendCoinsFromAccountToModule(ctx, address, types.ModuleName, feeCoins)
	if err != nil {
		return err
	}

	return nil
}
