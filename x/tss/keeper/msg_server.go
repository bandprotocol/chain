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

	// use LastCommitHash as a dkgContext
	dkgContext := ctx.BlockHeader().LastCommitHash
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
	memberID, found := k.GetMemberID(ctx, groupID, req.Member)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrMemberNotFound, "member id is not found")
	}

	// check previous commitment
	_, found = k.GetRound1Commitments(ctx, groupID, memberID)
	if found {
		return nil, sdkerrors.Wrap(types.ErrAlreadyCommitRound1, "this member already commit round 1 ")
	}

	// get dkg-context
	dkgContext, found := k.GetDKGContext(ctx, groupID)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrDKGContextNotFound, "dkg-context is not found")
	}

	valid, err := tss.VerifyOneTimeSig(req.GroupID, dkgContext, req.OneTimeSig, req.OneTimePubKey)
	if !valid || err != nil {
		return nil, sdkerrors.Wrap(types.ErrVerifyOneTimeSigFailed, err.Error())
	}

	valid, err = tss.VerifyA0Sig(req.GroupID, dkgContext, req.A0Sig, types.PublicKey(req.CoefficientsCommit[0]))
	if !valid || err != nil {
		return nil, sdkerrors.Wrap(types.ErrVerifyA0SigFailed, err.Error())
	}

	round1Commitments := types.Round1Commitments{
		CoefficientsCommit: req.CoefficientsCommit,
		OneTimePubKey:      req.OneTimePubKey,
		A0Sig:              req.A0Sig,
		OneTimeSig:         req.OneTimeSig,
	}
	k.SetRound1Commitments(ctx, groupID, memberID, round1Commitments)

	count := k.GetRound1CommitmentsCount(ctx, groupID)
	if count == group.Size_ {
		group.Status = types.ROUND_2
		k.UpdateGroup(ctx, groupID, group)
	}

	return &types.MsgSubmitDKGRound1Response{}, nil
}
