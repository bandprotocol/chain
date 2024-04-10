package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

func (k Keeper) CreateGroupReplacement(
	ctx sdk.Context,
	newGroupID tss.GroupID,
	execTime time.Time,
) (tss.SigningID, error) {
	currentGroupID := k.GetCurrentGroupID(ctx)
	currentGroup, err := k.tssKeeper.GetGroup(ctx, currentGroupID)
	if err != nil {
		return 0, err
	}

	replacement, err := k.GetReplacement(ctx)
	if err != nil {
		return 0, err
	}
	if replacement.SigningID != 0 {
		return 0, types.ErrReplacementInProgress
	}

	// create a signing request for the replacement
	newGroup, err := k.tssKeeper.GetGroup(ctx, newGroupID)
	if err != nil {
		return 0, err
	}
	msg := append(tsstypes.ReplaceGroupMsgPrefix, newGroup.PubKey...)
	signing, err := k.tssKeeper.CreateSigning(ctx, currentGroup, msg)
	if err != nil {
		return 0, err
	}

	k.SetReplacement(ctx, types.Replacement{
		SigningID: signing.ID,
		ExecTime:  execTime,
	})
	k.SetReplacingGroupID(ctx, newGroupID)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReplacement,
			sdk.NewAttribute(tsstypes.AttributeKeySigningID, fmt.Sprintf("%d", signing.ID)),
			sdk.NewAttribute(types.AttributeKeyCurrentGroupID, fmt.Sprintf("%d", currentGroupID)),
			sdk.NewAttribute(types.AttributeKeyReplacingGroupID, fmt.Sprintf("%d", newGroupID)),
		),
	)

	return signing.ID, nil
}

// HandleReplaceGroup updates the group information after a successful signing process.
func (k Keeper) HandleReplaceGroup(ctx sdk.Context, endBlockTime time.Time) error {
	replacement, err := k.GetReplacement(ctx)
	if err != nil {
		return err
	}
	if replacement.SigningID == 0 || endBlockTime.Before(replacement.ExecTime) {
		return nil
	}

	// Retrieve information about group.
	currentGroupID := k.GetCurrentGroupID(ctx)
	replacingGroupID := k.GetReplacingGroupID(ctx)

	// Retrieve information about signing.
	signing, err := k.tssKeeper.GetSigning(ctx, replacement.SigningID)
	if err != nil {
		return err
	}

	// If the signing process is unsuccessful, cleared data and emit event that fails to replace group.
	if signing.Status != tsstypes.SIGNING_STATUS_SUCCESS {
		k.SetReplacement(ctx, types.Replacement{})
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeReplacementFailed,
				sdk.NewAttribute(tsstypes.AttributeKeySigningID, fmt.Sprintf("%d", replacement.SigningID)),
				sdk.NewAttribute(types.AttributeKeyCurrentGroupID, fmt.Sprintf("%d", currentGroupID)),
				sdk.NewAttribute(types.AttributeKeyReplacingGroupID, fmt.Sprintf("%d", replacingGroupID)),
			),
		)
		return nil
	}

	// clear members from the current group and add members from the new group.
	oldMembers := k.tssKeeper.MustGetMembers(ctx, currentGroupID)
	for _, m := range oldMembers {
		k.DeleteMember(ctx, sdk.MustAccAddressFromBech32(m.Address))
	}

	k.SetCurrentGroupID(ctx, replacingGroupID)
	k.SetReplacingGroupID(ctx, tss.GroupID(0))

	newMembers := k.tssKeeper.MustGetMembers(ctx, replacingGroupID)
	for _, m := range newMembers {
		if err := k.AddNewMember(ctx, sdk.MustAccAddressFromBech32(m.Address)); err != nil {
			return err
		}
	}

	// clear replacement info and emit an event.
	k.SetReplacement(ctx, types.Replacement{})
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReplacementSuccess,
			sdk.NewAttribute(tsstypes.AttributeKeySigningID, fmt.Sprintf("%d", replacement.SigningID)),
			sdk.NewAttribute(types.AttributeKeyCurrentGroupID, fmt.Sprintf("%d", currentGroupID)),
			sdk.NewAttribute(types.AttributeKeyReplacingGroupID, fmt.Sprintf("%d", replacingGroupID)),
		),
	)

	return nil
}
