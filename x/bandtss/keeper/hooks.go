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
	for _, m := range members {
		h.k.SetActiveStatus(ctx, sdk.MustAccAddressFromBech32(m.Address))
	}

	h.k.SetCurrentGroupID(ctx, group.ID)
}

func (h Hooks) AfterCreatingGroupFailed(ctx sdk.Context, group tsstypes.Group) {}

func (h Hooks) BeforeSetGroupExpired(ctx sdk.Context, group tsstypes.Group) {
	// TODO: Penalize members will be slashed in the future.
}

func (h Hooks) AfterReplacingGroupCompleted(ctx sdk.Context, replacement tsstypes.Replacement) {
	// check if this signing is from the bandtss module
	// unlikely to get an error from GetGroup but log the error just in case.
	group, err := h.k.tssKeeper.GetGroup(ctx, replacement.CurrentGroupID)
	if err != nil {
		h.k.Logger(ctx).Error(fmt.Sprintf("Error getting groupID %v: %v", replacement.CurrentGroupID, err))
		return
	}
	if group.ModuleOwner != types.ModuleName {
		return
	}

	oldMembers := h.k.tssKeeper.MustGetMembers(ctx, replacement.CurrentGroupID)
	for _, m := range oldMembers {
		addr := sdk.AccAddress(m.PubKey)
		h.k.DeleteStatus(ctx, addr)
	}

	newMembers := h.k.tssKeeper.MustGetMembers(ctx, replacement.NewGroupID)
	for _, m := range newMembers {
		h.k.SetActiveStatus(ctx, sdk.MustAccAddressFromBech32(m.Address))
	}

	h.k.SetCurrentGroupID(ctx, replacement.NewGroupID)
	h.k.SetReplacingGroupID(ctx, tss.GroupID(0))
}

func (h Hooks) AfterReplacingGroupFailed(ctx sdk.Context, replacement tsstypes.Replacement) {
	// check if this signing is from the bandtss module
	// unlikely to get an error from GetGroup but log the error just in case.
	group, err := h.k.tssKeeper.GetGroup(ctx, replacement.CurrentGroupID)
	if err != nil {
		h.k.Logger(ctx).Error(fmt.Sprintf("Error getting groupID %v: %v", replacement.CurrentGroupID, err))
		return
	}
	if group.ModuleOwner != types.ModuleName {
		return
	}

	h.k.SetReplacingGroupID(ctx, tss.GroupID(0))
}

func (h Hooks) AfterSigningFailed(ctx sdk.Context, signing tsstypes.Signing) {
	// check if this signing is from the bandtss module
	// unlikely to get an error from GetGroup but log the error just in case.
	group, err := h.k.tssKeeper.GetGroup(ctx, signing.GroupID)
	if err != nil {
		h.k.Logger(ctx).Error(fmt.Sprintf("Error getting groupID %v: %v", signing.GroupID, err))
		return
	}
	if group.ModuleOwner != types.ModuleName {
		return
	}

	// refund fee to requester. Unlikely to get an error from refund fee, but log it just in case.
	if err := h.k.RefundFee(ctx, signing); err != nil {
		h.k.Logger(ctx).Error(fmt.Sprintf("Error refunding fee signingID %v : %v", signing.ID, err))
		return
	}

	h.k.DeleteSigningFee(ctx, signing.ID)
}

func (h Hooks) BeforeSetSigningExpired(ctx sdk.Context, signing tsstypes.Signing) {
	// check if this signing is from the bandtss module
	// unlikely to get an error from GetGroup but log the error just in case.
	group, err := h.k.tssKeeper.GetGroup(ctx, signing.GroupID)
	if err != nil {
		h.k.Logger(ctx).Error(fmt.Sprintf("Error getting groupID %v: %v", signing.GroupID, err))
		return
	}
	if group.ModuleOwner != types.ModuleName {
		return
	}

	penalizedMembers, err := h.k.tssKeeper.GetPenalizedMembersExpiredSigning(ctx, signing)
	// unlikely to get an error (convert to address type), but log the error just in case
	if err != nil {
		h.k.Logger(ctx).Error(fmt.Sprintf("Error getting penalized members: %v", err))
	}

	for _, addr := range penalizedMembers {
		h.k.SetInactiveStatus(ctx, addr)
	}

	// refund fee to requester. Unlikely to get an error from refund fee, but log it just in case.
	if err := h.k.RefundFee(ctx, signing); err != nil {
		h.k.Logger(ctx).Error(fmt.Sprintf("Error refunding fee signingID %v : %v", signing.ID, err))
		return
	}

	h.k.DeleteSigningFee(ctx, signing.ID)
}

func (h Hooks) AfterSigningCompleted(ctx sdk.Context, signing tsstypes.Signing) {
	// check if this signing is from the bandtss module
	// unlikely to get an error from GetGroup but log the error just in case.
	group, err := h.k.tssKeeper.GetGroup(ctx, signing.GroupID)
	if err != nil {
		h.k.Logger(ctx).Error(fmt.Sprintf("Error getting groupID %v: %v", signing.GroupID, err))
		return
	}
	if group.ModuleOwner != types.ModuleName {
		return
	}

	// get signing fee. Unlikely to get an error; but log the error just in case.
	signingFee, err := h.k.GetSigningFee(ctx, signing.ID)
	if err != nil {
		h.k.Logger(ctx).Error(fmt.Sprintf("Error getting signing fee: %v", err))
		return
	}

	// no fee is transferred, end process.
	if signingFee.Fee.IsZero() {
		return
	}

	// Send fee to assigned members.
	for _, am := range signing.AssignedMembers {
		address := sdk.MustAccAddressFromBech32(am.Address)

		// unlikely to get an error, but log the error just in case
		if err := h.k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, signingFee.Fee); err != nil {
			h.k.Logger(ctx).Error(fmt.Sprintf("Failed to send fee to address %s: %v", am.Address, err))
		}
	}
}
