package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tss/types"
)

var _ types.MsgServer = Keeper{}

func (k Keeper) CreateGroup(goCtx context.Context, req *types.MsgCreateGroup) (*types.MsgCreateGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// create new group
	groupID := k.CreateNewGroup(ctx, types.Group{
		Size_:     uint32(len(req.Members)),
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

	return &types.MsgCreateGroupResponse{}, nil
}
