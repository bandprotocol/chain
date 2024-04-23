package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// ClaimRewards defines a method for creating a new validator
func (k msgServer) ClaimRewards(
	goCtx context.Context,
	msg *types.MsgClaimRewards,
) (*types.MsgClaimRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, err
	}

	// update rewards from all stakes of an address
	stakes := k.GetStakes(ctx, address)
	for _, stake := range stakes {
		k.ProcessStake(ctx, stake)
	}

	// calculate claim rewards
	claimRewards := sdk.NewDecCoins()
	rewards := k.GetRewards(ctx, address)
	for _, reward := range rewards {
		claimRewards = claimRewards.Add(reward.Amounts...)
		k.DeleteReward(ctx, address, reward.Key)
	}

	// truncate reward dec coins, return remainder to community pool
	finalRewards, remainder := claimRewards.TruncateDecimal()

	// add coins to user account
	if !finalRewards.IsZero() {
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, finalRewards)
		if err != nil {
			return nil, err
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaimRewards,
			sdk.NewAttribute(types.AttributeKeyAddress, msg.Address),
			sdk.NewAttribute(sdk.AttributeKeyAmount, finalRewards.String()),
		),
	)

	k.addRemainderAmount(ctx, remainder)

	return &types.MsgClaimRewardsResponse{}, nil
}

func (k msgServer) LockToken(
	goCtx context.Context,
	msg *types.MsgLockToken,
) (*types.MsgLockTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, err
	}

	err = k.SetLockedToken(ctx, address, msg.Key, msg.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgLockTokenResponse{}, nil
}

func (k msgServer) AddRewards(
	goCtx context.Context,
	msg *types.MsgAddRewards,
) (*types.MsgAddRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	err = k.Keeper.AddRewards(ctx, sender, msg.Key, msg.Rewards)
	if err != nil {
		return nil, err
	}

	return &types.MsgAddRewardsResponse{}, nil
}

func (k msgServer) DeactivateKey(
	goCtx context.Context,
	msg *types.MsgDeactivateKey,
) (*types.MsgDeactivateKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := k.Keeper.DeactivateKey(ctx, msg.Key)
	if err != nil {
		return nil, err
	}

	return &types.MsgDeactivateKeyResponse{}, nil
}
