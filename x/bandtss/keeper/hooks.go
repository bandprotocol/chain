package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
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

func (h Hooks) AfterCreatingGroupCompleted(ctx sdk.Context, group tsstypes.Group) {
	if group.ModuleOwner != types.ModuleName || h.k.GetCurrentGroupID(ctx) != 0 {
		return
	}

	members := h.k.tssKeeper.MustGetMembers(ctx, group.ID)
	addresses := make([]sdk.AccAddress, 0, len(members))
	for _, m := range members {
		addr := sdk.AccAddress(m.PubKey)
		addresses = append(addresses, addr)
	}

	h.k.SetCurrentGroupID(ctx, group.ID)
	h.k.SetActiveStatuses(ctx, addresses)
}

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

	h.k.SetJailStatuses(ctx, penalizedMembers)
}

func (h Hooks) AfterReplacingGroupCompleted(ctx sdk.Context, replacement tsstypes.Replacement) {
	// check if this signing is from the bandtss module
	groupModule := h.k.tssKeeper.GetModuleOwner(ctx, replacement.CurrentGroupID)
	if groupModule != types.ModuleName {
		return
	}

	oldMembers := h.k.tssKeeper.MustGetMembers(ctx, replacement.CurrentGroupID)
	for _, m := range oldMembers {
		addr := sdk.AccAddress(m.PubKey)
		h.k.DeleteStatus(ctx, addr)
	}

	newMembers := h.k.tssKeeper.MustGetMembers(ctx, replacement.NewGroupID)
	addresses := make([]sdk.AccAddress, 0, len(newMembers))
	for _, m := range newMembers {
		addr := sdk.AccAddress(m.PubKey)
		addresses = append(addresses, addr)
	}

	h.k.SetActiveStatuses(ctx, addresses)
	h.k.SetCurrentGroupID(ctx, replacement.NewGroupID)
	h.k.SetReplacingGroupID(ctx, tss.GroupID(0))

}

func (h Hooks) AfterReplacingGroupFailed(ctx sdk.Context, replacement tsstypes.Replacement) {
	// check if this signing is from the bandtss module
	groupModule := h.k.tssKeeper.GetModuleOwner(ctx, replacement.CurrentGroupID)
	if groupModule != types.ModuleName {
		return
	}

	h.k.SetReplacingGroupID(ctx, tss.GroupID(0))
}

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

	h.k.SetInactiveStatuses(ctx, penalizedMembers)
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
	return nil
}

func (h Hooks) AfterPollDE(ctx sdk.Context, member sdk.AccAddress) error {
	return nil
}
