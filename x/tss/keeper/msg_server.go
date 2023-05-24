package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

var _ types.MsgServer = Keeper{}

// CreateGroup handles the request to create a new group.
// It first unwraps the Go context into an SDK context, then creates a new group with the given members.
// Afterwards, it sets each member into the KVStore, and hashes the groupID with the LastCommitHash from the block header to create the DKG context.
// Finally, it emits an event for the group creation.
func (k Keeper) CreateGroup(goCtx context.Context, req *types.MsgCreateGroup) (*types.MsgCreateGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	groupSize := uint64(len(req.Members))

	// Create new group
	groupID := k.CreateNewGroup(ctx, types.Group{
		Size_:     groupSize,
		Threshold: req.Threshold,
		PubKey:    nil,
		Status:    types.ROUND_1,
	})

	// Set members
	for i, m := range req.Members {
		// id start from 1
		k.SetMember(ctx, groupID, tss.MemberID(i+1), types.Member{
			Member: m,
			PubKey: tss.PublicKey(nil),
		})
	}

	// Use LastCommitHash and groupID to hash to dkgContext
	dkgContext := tss.Hash(sdk.Uint64ToBigEndian(uint64(groupID)), ctx.BlockHeader().LastCommitHash)
	k.SetDKGContext(ctx, groupID, dkgContext)

	event := sdk.NewEvent(
		types.EventTypeCreateGroup,
		sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
		sdk.NewAttribute(types.AttributeKeySize, fmt.Sprintf("%d", groupSize)),
		sdk.NewAttribute(types.AttributeKeyThreshold, fmt.Sprintf("%d", req.Threshold)),
		sdk.NewAttribute(types.AttributeKeyPubKey, ""),
		sdk.NewAttribute(types.AttributeKeyStatus, types.ROUND_1.String()),
		sdk.NewAttribute(types.AttributeKeyDKGContext, hex.EncodeToString(dkgContext)),
	)
	for _, m := range req.Members {
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyMember, m))
	}
	ctx.EventManager().EmitEvent(event)

	return &types.MsgCreateGroupResponse{}, nil
}

// SubmitDKGRound1 handles the submission of round1 in the DKG process.
// After unwrapping the context, it first checks the status of the group, and whether the member is valid and has not submitted before.
// Then, it retrieves the DKG context for the group and verifies the one-time signature and A0 signature.
// If all checks pass, it saves the round1 data into the KVStore and emits an event for the submission.
// If all members have submitted their round1 data, it updates the status of the group to round 2 and emits an event for the completion of round1.
func (k Keeper) SubmitDKGRound1(
	goCtx context.Context,
	req *types.MsgSubmitDKGRound1,
) (*types.MsgSubmitDKGRound1Response, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := req.GroupID
	memberID := req.Round1Data.MemberID

	// Check group status
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if group.Status != types.ROUND_1 {
		return nil, sdkerrors.Wrap(types.ErrRoundExpired, "group status is not round1")
	}

	// Verify member
	isMember := k.VerifyMember(ctx, groupID, memberID, req.Member)
	if !isMember {
		return nil, sdkerrors.Wrapf(
			types.ErrMemberNotAuthorized,
			"member %s is not in group %d",
			req.Member,
			groupID,
		)
	}

	// Check previous submit
	_, err = k.GetRound1Data(ctx, groupID, req.Round1Data.MemberID)
	if err == nil {
		return nil, sdkerrors.Wrap(types.ErrAlreadyCommitRound1, "this member already submit round1 ")
	}

	// Get dkg-context
	dkgContext, err := k.GetDKGContext(ctx, groupID)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrDKGContextNotFound, "dkg-context is not found")
	}

	err = tss.VerifyOneTimeSig(memberID, dkgContext, req.Round1Data.OneTimeSig, req.Round1Data.OneTimePubKey)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrVerifyOneTimeSigFailed, err.Error())
	}

	err = tss.VerifyA0Sig(
		memberID,
		dkgContext,
		req.Round1Data.A0Sig,
		tss.PublicKey(req.Round1Data.CoefficientsCommit[0]),
	)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrVerifyA0SigFailed, err.Error())
	}

	k.SetRound1Data(ctx, groupID, req.Round1Data)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitDKGRound1,
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", memberID)),
			sdk.NewAttribute(types.AttributeKeyMember, req.Member),
			sdk.NewAttribute(types.AttributeKeyCoefficientsCommit, req.Round1Data.CoefficientsCommit.ToString()),
			sdk.NewAttribute(types.AttributeKeyOneTimePubKey, hex.EncodeToString(req.Round1Data.OneTimePubKey)),
			sdk.NewAttribute(types.AttributeKeyA0Sig, hex.EncodeToString(req.Round1Data.A0Sig)),
			sdk.NewAttribute(types.AttributeKeyOneTimeSig, hex.EncodeToString(req.Round1Data.OneTimeSig)),
		),
	)

	count := k.GetRound1DataCount(ctx, groupID)
	if count == group.Size_ {
		group.Status = types.ROUND_2
		k.UpdateGroup(ctx, groupID, group)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRound1Success,
				sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
				sdk.NewAttribute(types.AttributeKeyStatus, group.Status.String()),
			),
		)
	}

	return &types.MsgSubmitDKGRound1Response{}, nil
}
