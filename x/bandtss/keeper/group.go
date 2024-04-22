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

	replacement := k.GetReplacement(ctx)
	if replacement.Status == types.REPLACEMENT_STATUS_WAITING_SIGNING ||
		replacement.Status == types.REPLACEMENT_STATUS_WAITING_REPLACE {
		return 0, types.ErrReplacementInProgress
	}

	// create a signing request for the replacement
	newGroup, err := k.tssKeeper.GetGroup(ctx, newGroupID)
	if err != nil {
		return 0, err
	}

	// Execute the handler to process the replacement request.
	msg, err := k.tssKeeper.HandleSigningContent(ctx, types.NewReplaceGroupSignatureOrder(newGroup.PubKey))
	if err != nil {
		return 0, err
	}
	signing, err := k.tssKeeper.CreateSigning(ctx, currentGroup, msg)
	if err != nil {
		return 0, err
	}

	k.SetReplacement(ctx, types.Replacement{
		SigningID:      signing.ID,
		CurrentGroupID: currentGroupID,
		CurrentPubKey:  currentGroup.PubKey,
		NewGroupID:     newGroupID,
		NewPubKey:      newGroup.PubKey,
		Status:         types.REPLACEMENT_STATUS_WAITING_SIGNING,
		ExecTime:       execTime,
	})

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
	replacement := k.GetReplacement(ctx)

	// check signing status and update replacement status.
	if replacement.Status == types.REPLACEMENT_STATUS_WAITING_SIGNING {
		signing, err := k.tssKeeper.GetSigning(ctx, replacement.SigningID)
		if err != nil {
			return err
		}

		if signing.IsFailed() ||
			(signing.Status == tsstypes.SIGNING_STATUS_WAITING && endBlockTime.After(replacement.ExecTime)) {
			return k.HandleFailReplacementSigning(ctx, replacement)
		}

		if signing.Status == tsstypes.SIGNING_STATUS_SUCCESS {
			replacement.Status = types.REPLACEMENT_STATUS_WAITING_REPLACE
			k.SetReplacement(ctx, replacement)
		}
	}

	if replacement.Status == types.REPLACEMENT_STATUS_WAITING_REPLACE && endBlockTime.After(replacement.ExecTime) {
		return k.ReplaceGroup(ctx, replacement)
	}

	return nil
}

func (k Keeper) HandleFailReplacementSigning(ctx sdk.Context, replacement types.Replacement) error {
	replacement.Status = types.REPLACEMENT_STATUS_FALLEN
	k.SetReplacement(ctx, replacement)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReplacementFailed,
			sdk.NewAttribute(tsstypes.AttributeKeySigningID, fmt.Sprintf("%d", replacement.SigningID)),
			sdk.NewAttribute(types.AttributeKeyCurrentGroupID, fmt.Sprintf("%d", replacement.CurrentGroupID)),
			sdk.NewAttribute(types.AttributeKeyReplacingGroupID, fmt.Sprintf("%d", replacement.NewGroupID)),
		),
	)
	return nil
}

func (k Keeper) ReplaceGroup(ctx sdk.Context, replacement types.Replacement) error {
	// clear members from the current group and add members from the new group.
	oldMembers := k.tssKeeper.MustGetMembers(ctx, replacement.CurrentGroupID)
	for _, m := range oldMembers {
		k.DeleteMember(ctx, sdk.MustAccAddressFromBech32(m.Address))
	}

	k.SetCurrentGroupID(ctx, replacement.NewGroupID)

	newMembers := k.tssKeeper.MustGetMembers(ctx, replacement.NewGroupID)
	for _, m := range newMembers {
		if err := k.AddNewMember(ctx, sdk.MustAccAddressFromBech32(m.Address)); err != nil {
			return err
		}
	}

	// update replacement status and emit an event.
	replacement.Status = types.REPLACEMENT_STATUS_SUCCESS
	k.SetReplacement(ctx, replacement)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReplacementSuccess,
			sdk.NewAttribute(tsstypes.AttributeKeySigningID, fmt.Sprintf("%d", replacement.SigningID)),
			sdk.NewAttribute(types.AttributeKeyCurrentGroupID, fmt.Sprintf("%d", replacement.CurrentGroupID)),
			sdk.NewAttribute(types.AttributeKeyReplacingGroupID, fmt.Sprintf("%d", replacement.NewGroupID)),
		),
	)

	return nil
}
