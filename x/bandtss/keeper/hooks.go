package keeper

import (
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

	// skip the process if the signing is for replacement
	replacement, err := h.k.GetReplacement(ctx)
	if err != nil {
		return err
	}
	if signing.ID == replacement.SigningID {
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

	replacement, err := h.k.GetReplacement(ctx)
	if err != nil {
		return err
	}

	bandtssSigningID := h.k.GetSigningIDMapping(ctx, signing.ID)
	if bandtssSigningID == 0 && signing.ID != replacement.SigningID {
		return types.ErrSigningNotFound
	}

	// penalize members who didn't submit their partial signatures.
	penalizedMembers, err := h.k.tssKeeper.GetPenalizedMembersExpiredSigning(ctx, signing)
	if err != nil {
		return err
	}

	for _, addr := range penalizedMembers {
		if err := h.k.DeactivateMember(ctx, addr); err != nil {
			return err
		}
	}

	// if the signing is for replacement, exit the hooks.
	if signing.ID == replacement.SigningID {
		return nil
	}

	// if it is a signing initiated from bandtss module, check if the fee should be refunded and
	// remove the id mapping.
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

	// if it is a signing for replacement, exit the hooks.
	replacement, err := h.k.GetReplacement(ctx)
	if err != nil {
		return err
	}
	if signing.ID == replacement.SigningID {
		return nil
	}

	bandtssSigningID := h.k.GetSigningIDMapping(ctx, signing.ID)
	if bandtssSigningID == 0 {
		return types.ErrSigningNotFound
	}

	bandtssSigning := h.k.MustGetSigning(ctx, bandtssSigningID)

	// Send fee to assigned members, if any.
	if signing.ID == bandtssSigning.CurrentGroupSigningID && !bandtssSigning.Fee.IsZero() {
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
	}

	h.k.DeleteSigningIDMapping(ctx, signing.ID)
	return nil
}
