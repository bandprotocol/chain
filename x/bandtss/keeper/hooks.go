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

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFirstGroupCreated,
			sdk.NewAttribute(types.AttributeKeyCurrentGroupID, fmt.Sprintf("%d", group.ID)),
		),
	)

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
	// Skip the process if the bandtssSigningID is not found in the mapping. If the signing
	// is for replacement, it does not have the bandtssSigningID, and no refund is required.
	bandtssSigningID := h.k.GetSigningIDMapping(ctx, signing.ID)
	if bandtssSigningID == 0 {
		return nil
	}

	// refund fee to requester. Unlikely to get an error from refund fee, but log it just in case.
	if err := h.k.CheckRefundFee(ctx, signing, bandtssSigningID); err != nil {
		return err
	}

	h.k.DeleteSigningIDMapping(ctx, signing.ID)
	return nil
}

func (h Hooks) BeforeSetSigningExpired(ctx sdk.Context, signing tsstypes.Signing) error {
	// penalize members who didn't submit their partial signatures if the signing is from
	// the current group (at the moment).
	currentGroupID := h.k.GetCurrentGroupID(ctx)
	if signing.GroupID == currentGroupID {
		penalizedMembers, err := h.k.tssKeeper.GetPenalizedMembersExpiredSigning(ctx, signing)
		if err != nil {
			return err
		}

		for _, addr := range penalizedMembers {
			if err := h.k.DeactivateMember(ctx, addr); err != nil {
				return err
			}
		}
	}

	// Skip the process if the bandtssSigningID is not found in the mapping. If the signing
	// is for replacement, it does not have the bandtssSigningID, and no refund is required.
	bandtssSigningID := h.k.GetSigningIDMapping(ctx, signing.ID)
	if bandtssSigningID == 0 {
		return nil
	}

	// refund fee to requester and delete the signingID mapping.
	if err := h.k.CheckRefundFee(ctx, signing, bandtssSigningID); err != nil {
		return err
	}
	h.k.DeleteSigningIDMapping(ctx, signing.ID)

	return nil
}

func (h Hooks) AfterSigningCompleted(ctx sdk.Context, signing tsstypes.Signing) error {
	// Skip the process if the bandtssSigningID is not found in the mapping. If the signing
	// is for replacement, it does not have the bandtssSigningID, and no fee transfer is required.
	bandtssSigningID := h.k.GetSigningIDMapping(ctx, signing.ID)
	if bandtssSigningID == 0 {
		return nil
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
