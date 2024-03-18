package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// Wrapper struct
type Hooks struct {
	k Keeper
}

var _ tsstypes.TSSHooks = Hooks{}

// Create new Bandtss hooks
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) AfterCreatingGroupCompleted(ctx sdk.Context, group tsstypes.Group) {}

func (h Hooks) AfterCreatingGroupFailed(ctx sdk.Context, group tsstypes.Group) {}

func (h Hooks) BeforeSetGroupExpired(ctx sdk.Context, group tsstypes.Group) {
	if group.ModuleOwner != types.ModuleName {
		return
	}

	penalizedMembers, err := h.k.tssKeeper.GetPenalizedMembersExpiredGroup(ctx, group)
	// error is from we cannot find groupID in the store. In this case, we don't need to do anything,
	// but log the error just in case.
	if err != nil {
		h.k.Logger(ctx).Error(fmt.Sprintf("Error getting penalized members: %v", err))
		return
	}

	for _, m := range penalizedMembers {
		h.k.SetJailStatus(ctx, m)
	}
}

func (h Hooks) AfterReplacingGroupCompleted(ctx sdk.Context, replacement tsstypes.Replacement) {}

func (h Hooks) AfterReplacingGroupFailed(ctx sdk.Context, replacement tsstypes.Replacement) {}

func (h Hooks) AfterSigningFailed(ctx sdk.Context, signing tsstypes.Signing) {
	if signing.Fee.IsZero() {
		return
	}

	// check if this signing is from the bandtss module
	groupModule := h.k.tssKeeper.GetModuleOwner(ctx, signing.GroupID)
	if groupModule != types.ModuleName {
		return
	}

	// Refund fee to requester
	address := sdk.MustAccAddressFromBech32(signing.Requester)
	feeCoins := signing.Fee.MulInt(sdk.NewInt(int64(len(signing.AssignedMembers))))

	err := h.k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, feeCoins)
	// unlikely to get an error, but log the error just in case
	if err != nil {
		h.k.Logger(ctx).Error(fmt.Sprintf("Failed to refund fee to address %s: %v", signing.Requester, err))
	}
}

func (h Hooks) BeforeSetSigningExpired(ctx sdk.Context, signing tsstypes.Signing) {
	// check if this signing is from the bandtss module
	groupModule := h.k.tssKeeper.GetModuleOwner(ctx, signing.GroupID)
	if groupModule != types.ModuleName {
		return
	}

	penalizedMembers, err := h.k.tssKeeper.GetPenalizedMembersExpiredSigning(ctx, signing)
	// unlikely to get an error (convert to address type), but log the error just in case
	if err != nil {
		h.k.Logger(ctx).Error(fmt.Sprintf("Error getting penalized members: %v", err))
	}

	for _, m := range penalizedMembers {
		h.k.SetInactiveStatus(ctx, m)
	}
}

func (h Hooks) AfterSigningCompleted(ctx sdk.Context, signing tsstypes.Signing) {
	// check if this signing is from the bandtss module
	groupModule := h.k.tssKeeper.GetModuleOwner(ctx, signing.GroupID)
	if groupModule != types.ModuleName {
		return
	}

	// Send fee to assigned members.
	for _, am := range signing.AssignedMembers {
		address := sdk.MustAccAddressFromBech32(am.Address)

		// unlikely to get an error, but log the error just in case
		if err := h.k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, signing.Fee); err != nil {
			h.k.Logger(ctx).Error(fmt.Sprintf("Failed to send fee to address %s: %v", am.Address, err))
		}
	}
}

func (h Hooks) AfterSigningCreated(ctx sdk.Context, signing tsstypes.Signing) error {
	// check if this signing is from the bandtss module
	groupModule := h.k.tssKeeper.GetModuleOwner(ctx, signing.GroupID)
	if groupModule != types.ModuleName {
		return nil
	}

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

func (h Hooks) AfterHandleSetDEs(ctx sdk.Context, address sdk.AccAddress) error {
	// only update status if the member was paused
	status := h.k.GetStatus(ctx, address)
	if status.Status != types.MEMBER_STATUS_PAUSED {
		return nil
	}

	// if DE is still empty, keep its status as is.
	left := h.k.tssKeeper.GetDECount(ctx, address)
	if left == 0 {
		return nil
	}

	// Set status to active and update the status in tssKeeper
	if err := h.k.SetActiveStatus(ctx, address); err != nil {
		return err
	}

	return nil
}

func (h Hooks) AfterPollDE(ctx sdk.Context, member sdk.AccAddress) error {
	left := h.k.tssKeeper.GetDECount(ctx, member)
	if left == 0 {
		h.k.SetPausedStatus(ctx, member)
	}

	return nil
}
