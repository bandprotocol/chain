package keeper

import (
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

func (h Hooks) AfterCreatingGroupCompleted(ctx sdk.Context, group tsstypes.Group) error {
	// check if this group is from the bandtss module or current group hasn't been set.
	if group.ModuleOwner != types.ModuleName || h.k.GetCurrentGroupID(ctx) != 0 {
		return nil
	}

	h.k.SetCurrentGroupID(ctx, group.ID)

	members := h.k.tssKeeper.MustGetMembers(ctx, group.ID)
	for _, m := range members {
		addr := sdk.MustAccAddressFromBech32(m.Address)
		if err := h.k.AddNewMember(ctx, addr); err != nil {
			return err
		}
	}

	return nil
}

func (h Hooks) AfterCreatingGroupFailed(ctx sdk.Context, group tsstypes.Group) error {
	return nil
}

func (h Hooks) BeforeSetGroupExpired(ctx sdk.Context, group tsstypes.Group) error {
	// TODO: Penalize members will be slashed in the future.
	return nil
}

func (h Hooks) AfterReplacingGroupCompleted(ctx sdk.Context, replacement tsstypes.Replacement) error {
	// check if this signing is from the bandtss module
	group, err := h.k.tssKeeper.GetGroup(ctx, replacement.CurrentGroupID)
	if err != nil {
		return err
	}
	if group.ModuleOwner != types.ModuleName {
		return nil
	}

	oldMembers := h.k.tssKeeper.MustGetMembers(ctx, replacement.CurrentGroupID)
	for _, m := range oldMembers {
		h.k.DeleteMember(ctx, sdk.MustAccAddressFromBech32(m.Address))
	}

	h.k.SetCurrentGroupID(ctx, replacement.NewGroupID)
	h.k.SetReplacingGroupID(ctx, tss.GroupID(0))

	newMembers := h.k.tssKeeper.MustGetMembers(ctx, replacement.NewGroupID)
	for _, m := range newMembers {
		if err := h.k.AddNewMember(ctx, sdk.MustAccAddressFromBech32(m.Address)); err != nil {
			return err
		}
	}

	return nil
}

func (h Hooks) AfterReplacingGroupFailed(ctx sdk.Context, replacement tsstypes.Replacement) error {
	// check if this signing is from the bandtss module
	group, err := h.k.tssKeeper.GetGroup(ctx, replacement.CurrentGroupID)
	if err != nil {
		return err
	}

	if group.ModuleOwner != types.ModuleName {
		return nil
	}

	h.k.SetReplacingGroupID(ctx, tss.GroupID(0))
	return nil
}

func (h Hooks) AfterSigningFailed(ctx sdk.Context, signing tsstypes.Signing) error {
	// check if this signing is from the bandtss module
	// unlikely to get an error from GetGroup but log the error just in case.
	group, err := h.k.tssKeeper.GetGroup(ctx, signing.GroupID)
	if err != nil {
		return err
	}
	if group.ModuleOwner != types.ModuleName {
		return nil
	}

	bandtssSigningID := h.k.GetSigningIDMapping(ctx, signing.ID)
	if bandtssSigningID == 0 {
		return types.ErrSigningNotFound
	}

	// refund fee to requester. Unlikely to get an error from refund fee, but log it just in case.
	if err := h.k.CheckRefundFee(ctx, signing); err != nil {
		return err
	}

	h.k.DeleteSigningIDMapping(ctx, signing.ID)
	return nil
}

func (h Hooks) BeforeSetSigningExpired(ctx sdk.Context, signing tsstypes.Signing) error {
	// check if this signing is from the bandtss module
	group, err := h.k.tssKeeper.GetGroup(ctx, signing.GroupID)
	if err != nil {
		return err
	}
	if group.ModuleOwner != types.ModuleName {
		return nil
	}

	bandtssSigningID := h.k.GetSigningIDMapping(ctx, signing.ID)
	if bandtssSigningID == 0 {
		return types.ErrSigningNotFound
	}

	penalizedMembers, err := h.k.tssKeeper.GetPenalizedMembersExpiredSigning(ctx, signing)
	if err != nil {
		return err
	}

	for _, addr := range penalizedMembers {
		if err := h.k.DeactivateMember(ctx, addr); err != nil {
			return err
		}
	}

	// check if the module should refund fee to requester.
	if err := h.k.CheckRefundFee(ctx, signing); err != nil {
		return err
	}

	h.k.DeleteSigningIDMapping(ctx, signing.ID)
	return nil
}

func (h Hooks) AfterSigningCompleted(ctx sdk.Context, signing tsstypes.Signing) error {
	// check if this signing is from the bandtss module
	group, err := h.k.tssKeeper.GetGroup(ctx, signing.GroupID)
	if err != nil {
		return err
	}
	if group.ModuleOwner != types.ModuleName {
		return nil
	}

	bandtssSigningID := h.k.GetSigningIDMapping(ctx, signing.ID)
	if bandtssSigningID == 0 {
		return types.ErrSigningNotFound
	}

	bandtssSigning := h.k.MustGetSigning(ctx, bandtssSigningID)

	// no fee is transferred, end process.
	if bandtssSigning.Fee.IsZero() || signing.ID != bandtssSigning.CurrentGroupSigningID {
		return nil
	}

	// Send fee to assigned members.
	for _, am := range signing.AssignedMembers {
		address := sdk.MustAccAddressFromBech32(am.Address)

		if err := h.k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			address,
			bandtssSigning.Fee,
		); err != nil {
			return err
		}
	}

	return nil
}
