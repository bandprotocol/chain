package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

var _ tsstypes.TSSCallback = &TSSCallback{}

// Wrapper struct
type TSSCallback struct {
	k Keeper
}

func NewTSSCallback(k Keeper) TSSCallback {
	return TSSCallback{k}
}

func (cb TSSCallback) OnGroupCreationCompleted(ctx sdk.Context, groupID tss.GroupID) {
	transition, found := cb.k.GetGroupTransition(ctx)
	if !found ||
		transition.IncomingGroupID != groupID ||
		transition.Status != types.TRANSITION_STATUS_CREATING_GROUP ||
		transition.ExecTime.Before(ctx.BlockTime()) {
		return
	}

	group := cb.k.tssKeeper.MustGetGroup(ctx, groupID)
	transition.IncomingGroupPubKey = group.PubKey

	// if the current group is not set, the transition is forced; update status and set
	// member into the group. Otherwise, create a signing request for transition.
	if transition.CurrentGroupID == 0 {
		// This shouldn't return error as the group is newly created.
		if err := cb.k.AddMembers(ctx, group.ID); err != nil {
			panic(err)
		}

		transition.Status = types.TRANSITION_STATUS_WAITING_EXECUTION
	} else {
		// create a signing request for transition. If the signing request is failed, set the
		// transition status to fallen.
		signingID, err := cb.k.CreateTransitionSigning(ctx, group.PubKey)
		if err != nil {
			cb.k.EndGroupTransitionProcess(ctx, transition, false)
			return
		}

		transition.Status = types.TRANSITION_STATUS_WAITING_SIGN
		transition.SigningID = signingID
	}

	// update the transition status and info.
	cb.k.SetGroupTransition(ctx, transition)

	// emit an event for the group transition.
	attrs := cb.k.ExtractEventAttributesFromTransition(transition)
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeGroupTransition, attrs...))
}

func (cb TSSCallback) OnGroupCreationFailed(ctx sdk.Context, groupID tss.GroupID) {
	transition, found := cb.k.GetGroupTransition(ctx)
	if found &&
		transition.IncomingGroupID == groupID &&
		transition.Status == types.TRANSITION_STATUS_CREATING_GROUP {
		cb.k.EndGroupTransitionProcess(ctx, transition, false)
	}
}

func (cb TSSCallback) OnGroupCreationExpired(ctx sdk.Context, groupID tss.GroupID) {
	transition, found := cb.k.GetGroupTransition(ctx)
	if found &&
		transition.IncomingGroupID == groupID &&
		transition.Status == types.TRANSITION_STATUS_CREATING_GROUP {
		cb.k.EndGroupTransitionProcess(ctx, transition, false)
	}
}

func (cb TSSCallback) OnSigningFailed(ctx sdk.Context, signingID tss.SigningID) {
	// If it is normal signing request (not for transition), delete the bandtssSigningID mapping.
	bandtssSigningID := cb.k.GetSigningIDMapping(ctx, signingID)
	if bandtssSigningID != 0 {
		cb.k.DeleteSigningIDMapping(ctx, signingID)
		return
	}

	// If the signing is for transition, update the transition status.
	transition, found := cb.k.GetGroupTransition(ctx)
	if found && signingID == transition.SigningID && transition.Status == types.TRANSITION_STATUS_WAITING_SIGN {
		cb.k.EndGroupTransitionProcess(ctx, transition, false)
	}
}

// OnSignTimeout is called when the signing request is expired. It penalizes the members that
// do not sign a message.
func (cb TSSCallback) OnSigningTimeout(
	ctx sdk.Context,
	signingID tss.SigningID,
	idleMembers []sdk.AccAddress,
) {
	signing := cb.k.tssKeeper.MustGetSigning(ctx, signingID)
	for _, addr := range idleMembers {
		// prevent deactivate member if the member is already deactivated or signing is from
		// previous current group.
		member, err := cb.k.GetMember(ctx, addr, signing.GroupID)
		if err != nil || !member.IsActive {
			continue
		}

		// Deactivate the member; this shouldn't cause an error because member should exists in
		// both tss and bandtss module.
		if err := cb.k.DeactivateMember(ctx, addr, signing.GroupID); err != nil {
			panic(err)
		}
	}
}

func (cb TSSCallback) OnSigningCompleted(
	ctx sdk.Context,
	signingID tss.SigningID,
	assignedMembers []sdk.AccAddress,
) {
	// If it is normal signing request (not for transition), transfer fee if needed and
	// delete the bandtssSigningID mapping.
	bandtssSigningID := cb.k.GetSigningIDMapping(ctx, signingID)
	if bandtssSigningID != 0 {
		bandtssSigning := cb.k.MustGetSigning(ctx, bandtssSigningID)
		cb.k.DeleteSigningIDMapping(ctx, signingID)

		// Send fee to assigned members, if any.
		if signingID != bandtssSigning.CurrentGroupSigningID || bandtssSigning.FeePerSigner.IsZero() {
			return
		}

		for _, addr := range assignedMembers {
			err := cb.k.bankKeeper.SendCoinsFromModuleToAccount(
				ctx,
				types.ModuleName,
				addr,
				bandtssSigning.FeePerSigner,
			)
			// It shouldn't return error as the fee is already transferred to the module account.
			if err != nil {
				panic(err)
			}
		}
		return
	}

	// If the signing is for transition, update the transition status.
	transition, found := cb.k.GetGroupTransition(ctx)
	if found && signingID == transition.SigningID && transition.Status == types.TRANSITION_STATUS_WAITING_SIGN {
		// add Members to the group, this shouldn't return error as the group is new to the module.
		if err := cb.k.AddMembers(ctx, transition.IncomingGroupID); err != nil {
			panic(err)
		}

		// update the transition status and info.
		transition.Status = types.TRANSITION_STATUS_WAITING_EXECUTION
		cb.k.SetGroupTransition(ctx, transition)

		// get signature and rAddress; this shouldn't return error as the signingID is already completed.
		signingResult, err := cb.k.tssKeeper.GetSigningResult(ctx, signingID)
		if err != nil {
			panic(err)
		}

		// emit an event for the group transition.
		attrs := cb.k.ExtractEventAttributesFromTransition(transition)
		attrs = append(attrs,
			sdk.NewAttribute(types.AttributeKeyRandomAddress, signingResult.EVMSignature.RAddress.String()),
			sdk.NewAttribute(types.AttributeKeySignature, signingResult.EVMSignature.Signature.String()),
		)

		ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeGroupTransition, attrs...))
	}
}
