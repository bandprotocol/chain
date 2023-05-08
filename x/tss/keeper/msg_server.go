package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

var _ types.MsgServer = Keeper{}

func (k Keeper) CreateGroup(goCtx context.Context, req *types.MsgCreateGroup) (*types.MsgCreateGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	groupSize := uint32(len(req.Members))

	// create new group
	groupID := k.CreateNewGroup(ctx, types.Group{
		Size_:     groupSize,
		Threshold: req.Threshold,
		PubKey:    nil,
		Status:    types.ROUND_1,
	})

	// set members
	for i, member := range req.Members {
		k.SetMember(ctx, groupID, uint64(i), types.Member{
			Signer: member,
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
	for _, member := range req.Members {
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyMember, member))
	}
	ctx.EventManager().EmitEvent(event)

	return &types.MsgCreateGroupResponse{}, nil
}

func (k Keeper) SubmitDKGRound1(
	goCtx context.Context,
	req *types.MsgSubmitDKGRound1,
) (*types.MsgSubmitDKGRound1Response, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO-CYLINDER: handle round1 msg
	fmt.Printf("req %+v \n", req)

	dkgContext := k.GetDKGContext(ctx, uint64(req.GroupID))

	valid, err := tss.VerifyOneTimeSig(req.GroupID, dkgContext, req.OneTimeSig, req.OneTimePubKey)
	if err != nil {
		return nil, err
	}

	fmt.Printf("result: %+v\n", valid)

	valid, err = tss.VerifyA0Sig(req.GroupID, dkgContext, req.A0Sig, types.PublicKey(req.CoefficientsCommit[0]))
	if err != nil {
		return nil, err
	}

	fmt.Printf("result: %+v\n", valid)

	return &types.MsgSubmitDKGRound1Response{}, nil
}
