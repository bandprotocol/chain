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

	key, err := k.GetKey(ctx, msg.Key)
	if err != nil {
		return nil, err
	}

	stake, err := k.GetStake(ctx, address, msg.Key)
	if err != nil {
		return nil, err
	}

	totalRewards := k.getTotalRewards(ctx, stake)
	truncatedTotalRewards, remainders := totalRewards.TruncateDecimal()
	finalRewards := truncatedTotalRewards.Add(stake.NegRewardDebts...).Sub(stake.PosRewardDebts...)

	if !finalRewards.IsZero() {
		stake.PosRewardDebts = truncatedTotalRewards
		stake.NegRewardDebts = sdk.NewCoins()
		k.SetStake(ctx, stake)

		err = k.bankKeeper.SendCoins(ctx, sdk.MustAccAddressFromBech32(key.Address), address, finalRewards)
		if err != nil {
			return nil, err
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeClaimRewards,
				sdk.NewAttribute(types.AttributeKeyAddress, msg.Address),
				sdk.NewAttribute(types.AttributeKeyKey, stake.Key),
				sdk.NewAttribute(sdk.AttributeKeyAmount, finalRewards.String()),
			),
		)
	}

	if !key.IsActive {
		k.DeleteStake(ctx, address, key.Name)

		key.Remainders = key.Remainders.Add(remainders...)
		k.SetKey(ctx, key)
	}

	return &types.MsgClaimRewardsResponse{}, nil
}

func (k msgServer) LockPower(
	goCtx context.Context,
	msg *types.MsgLockPower,
) (*types.MsgLockPowerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, err
	}

	err = k.SetLockedPower(ctx, address, msg.Key, msg.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgLockPowerResponse{}, nil
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
