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

func (k Keeper) CreateGroup(goCtx context.Context, req *types.MsgCreateGroup) (*types.MsgCreateGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	groupSize := uint64(len(req.Members))

	// create new group
	groupID := k.CreateNewGroup(ctx, types.Group{
		Size_:     groupSize,
		Threshold: req.Threshold,
		PubKey:    nil,
		Status:    types.ROUND_1,
	})

	// set members
	for i, m := range req.Members {
		k.SetMember(ctx, groupID, uint64(i), types.Member{
			Signer: m,
			PubKey: "",
		})
	}

	// use LastCommitHash and groupID to hash to dkgContext
	dkgContext := tss.Hash(sdk.Uint64ToBigEndian(groupID), ctx.BlockHeader().LastCommitHash)
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

func (k Keeper) SubmitDKGRound1(
	goCtx context.Context,
	req *types.MsgSubmitDKGRound1,
) (*types.MsgSubmitDKGRound1Response, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := uint64(req.GroupID)

	// check group status
	group, found := k.GetGroup(ctx, groupID)
	if !found {
		return nil, types.ErrGroupNotFound
	}

	if group.Status != types.ROUND_1 {
		return nil, sdkerrors.Wrap(types.ErrRound1AlreadyExpired, "group status is not round1")
	}

	// check members
	isMember := k.VerifyMember(ctx, groupID, uint64(req.MemberID), req.Member)
	if !isMember {
		return nil, sdkerrors.Wrap(
			types.ErrMemberNotFound,
			fmt.Sprintf("address: %s is not the member of this group", req.Member),
		)
	}

	// check previous commitment
	_, found = k.GetRound1Commitments(ctx, groupID, uint64(req.MemberID))
	if found {
		return nil, sdkerrors.Wrap(types.ErrAlreadyCommitRound1, "this member already commit round 1 ")
	}

	// get dkg-context
	dkgContext, found := k.GetDKGContext(ctx, groupID)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrDKGContextNotFound, "dkg-context is not found")
	}

	err := tss.VerifyOneTimeSig(req.GroupID, dkgContext, req.OneTimeSig, req.OneTimePubKey)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrVerifyOneTimeSigFailed, err.Error())
	}

	err = tss.VerifyA0Sig(req.GroupID, dkgContext, req.A0Sig, tss.PublicKey(req.CoefficientsCommit[0]))
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrVerifyA0SigFailed, err.Error())
	}

	round1Commitments := types.Round1Commitments{
		CoefficientsCommit: req.CoefficientsCommit,
		OneTimePubKey:      req.OneTimePubKey,
		A0Sig:              req.A0Sig,
		OneTimeSig:         req.OneTimeSig,
	}
	k.SetRound1Commitments(ctx, groupID, uint64(req.MemberID), round1Commitments)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypSubmitDKGRound1,
			sdk.NewAttribute(types.AttributeKeyCoefficientsCommit, round1Commitments.CoefficientsCommit.ToString()),
			sdk.NewAttribute(types.AttributeKeyOneTimePubKey, hex.EncodeToString(round1Commitments.OneTimePubKey)),
			sdk.NewAttribute(types.AttributeKeyA0Sig, hex.EncodeToString(round1Commitments.A0Sig)),
			sdk.NewAttribute(types.AttributeKeyOneTimeSig, hex.EncodeToString(round1Commitments.OneTimeSig)),
		),
	)

	count := k.GetRound1CommitmentsCount(ctx, groupID)
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
