package keeper

import (
	"fmt"
	"time"

	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// CreateGroupReplacement creates a replacement request to replace a current group ID with a new one
// which must be approved, i.e. request must be signed, from the current group.
func (k Keeper) CreateGroupReplacement(
	ctx sdk.Context,
	newGroupID tss.GroupID,
	execTime time.Time,
) (tss.SigningID, error) {
	if execTime.Before(ctx.BlockTime()) {
		return 0, types.ErrInvalidExecTime
	}

	currentGroupID := k.GetCurrentGroupID(ctx)
	if currentGroupID == 0 {
		return 0, types.ErrNoActiveGroup
	}

	currentGroup, err := k.tssKeeper.GetGroup(ctx, currentGroupID)
	if err != nil {
		return 0, err
	}

	replacement := k.GetReplacement(ctx)
	if replacement.Status == types.REPLACEMENT_STATUS_WAITING_SIGN ||
		replacement.Status == types.REPLACEMENT_STATUS_WAITING_REPLACE {
		return 0, types.ErrReplacementInProgress
	}

	// create a signing request for the replacement
	newGroup, err := k.tssKeeper.GetGroup(ctx, newGroupID)
	if err != nil {
		return 0, err
	}
	if newGroup.Status != tsstypes.GROUP_STATUS_ACTIVE {
		return 0, tsstypes.ErrGroupIsNotActive
	}

	// Execute the handler to process the replacement request.
	msg, err := k.tssKeeper.ConvertContentToBytes(ctx, types.NewReplaceGroupSignatureOrder(newGroup.PubKey))
	if err != nil {
		return 0, err
	}
	signing, err := k.tssKeeper.CreateSigning(ctx, currentGroup, msg)
	if err != nil {
		return 0, err
	}

	replacement = types.Replacement{
		SigningID:      signing.ID,
		CurrentGroupID: currentGroupID,
		CurrentPubKey:  currentGroup.PubKey,
		NewGroupID:     newGroupID,
		NewPubKey:      newGroup.PubKey,
		Status:         types.REPLACEMENT_STATUS_WAITING_SIGN,
		ExecTime:       execTime,
	}
	k.SetReplacement(ctx, replacement)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReplacement,
			sdk.NewAttribute(tsstypes.AttributeKeySigningID, fmt.Sprintf("%d", replacement.SigningID)),
			sdk.NewAttribute(types.AttributeKeyCurrentGroupID, fmt.Sprintf("%d", replacement.CurrentGroupID)),
			sdk.NewAttribute(types.AttributeKeyReplacingGroupID, fmt.Sprintf("%d", replacement.NewGroupID)),
			sdk.NewAttribute(types.AttributeKeyReplacementStatus, replacement.Status.String()),
			sdk.NewAttribute(types.AttributeKeyExecTime, replacement.ExecTime.Format(time.RFC3339)),
		),
	)

	return signing.ID, nil
}

// HandleReplaceGroup updates the replacement status or update the current group information after
// passing the replacement execution time and signing is completed. It emits an event at the end.
func (k Keeper) HandleReplaceGroup(ctx sdk.Context, endBlockTime time.Time) error {
	replacement := k.GetReplacement(ctx)

	// check signing status and update replacement status.
	if replacement.Status == types.REPLACEMENT_STATUS_WAITING_SIGN {
		signing, err := k.tssKeeper.GetSigning(ctx, replacement.SigningID)
		if err != nil {
			return err
		}

		if signing.Status == tsstypes.SIGNING_STATUS_EXPIRED ||
			signing.Status == tsstypes.SIGNING_STATUS_FALLEN ||
			(signing.Status == tsstypes.SIGNING_STATUS_WAITING && endBlockTime.After(replacement.ExecTime)) {
			return k.HandleFailReplacementSigning(ctx, replacement)
		}

		if signing.Status == tsstypes.SIGNING_STATUS_SUCCESS {
			replacement.Status = types.REPLACEMENT_STATUS_WAITING_REPLACE
			k.SetReplacement(ctx, replacement)

			newGroup, err := k.tssKeeper.GetGroup(ctx, replacement.NewGroupID)
			if err != nil {
				return err
			}

			rAddress, err := signing.Signature.R().Address()
			if err != nil {
				return err
			}
			sig := tmbytes.HexBytes(signing.Signature.S())

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeReplacement,
					sdk.NewAttribute(tsstypes.AttributeKeySigningID, fmt.Sprintf("%d", replacement.SigningID)),
					sdk.NewAttribute(types.AttributeKeyCurrentGroupID, fmt.Sprintf("%d", replacement.CurrentGroupID)),
					sdk.NewAttribute(types.AttributeKeyReplacingGroupID, fmt.Sprintf("%d", replacement.NewGroupID)),
					sdk.NewAttribute(types.AttributeKeyReplacementStatus, replacement.Status.String()),
					sdk.NewAttribute(types.AttributeKeyExecTime, replacement.ExecTime.Format(time.RFC3339)),
					sdk.NewAttribute(types.AttributeKeyNewGroupPubKey, newGroup.PubKey.String()),
					sdk.NewAttribute(types.AttributeKeyRAddress, tmbytes.HexBytes(rAddress).String()),
					sdk.NewAttribute(types.AttributeKeySignature, sig.String()),
				),
			)
		}
	}

	if replacement.Status == types.REPLACEMENT_STATUS_WAITING_REPLACE && endBlockTime.After(replacement.ExecTime) {
		return k.ReplaceGroup(ctx, replacement)
	}

	return nil
}

// HandleFailReplacementSigning update replacement status and emits an event.
func (k Keeper) HandleFailReplacementSigning(ctx sdk.Context, replacement types.Replacement) error {
	replacement.Status = types.REPLACEMENT_STATUS_FALLEN
	k.SetReplacement(ctx, replacement)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReplacement,
			sdk.NewAttribute(tsstypes.AttributeKeySigningID, fmt.Sprintf("%d", replacement.SigningID)),
			sdk.NewAttribute(types.AttributeKeyCurrentGroupID, fmt.Sprintf("%d", replacement.CurrentGroupID)),
			sdk.NewAttribute(types.AttributeKeyReplacingGroupID, fmt.Sprintf("%d", replacement.NewGroupID)),
			sdk.NewAttribute(types.AttributeKeyReplacementStatus, replacement.Status.String()),
			sdk.NewAttribute(types.AttributeKeyExecTime, replacement.ExecTime.Format(time.RFC3339)),
		),
	)
	return nil
}

// ReplaceGroup handle group replacement which includes manage members of the old and new group.
// and set new group ID to be a current one. It emits an event at the end.
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

	newGroup, err := k.tssKeeper.GetGroup(ctx, replacement.NewGroupID)
	if err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNewGroupActivate,
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", newGroup.ID)),
			sdk.NewAttribute(types.AttributeKeyGroupPubKey, newGroup.PubKey.String()),
		),
	)

	return nil
}
