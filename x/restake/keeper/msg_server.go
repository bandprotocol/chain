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

	address, err := sdk.AccAddressFromBech32(msg.LockerAddress)
	if err != nil {
		return nil, err
	}

	key, err := k.GetKey(ctx, msg.Key)
	if err != nil {
		return nil, err
	}

	lock, err := k.GetLock(ctx, address, msg.Key)
	if err != nil {
		return nil, err
	}

	totalRewards := k.getTotalRewards(ctx, lock)
	truncatedTotalRewards, remainders := totalRewards.TruncateDecimal()
	finalRewards := truncatedTotalRewards.Add(lock.NegRewardDebts...).Sub(lock.PosRewardDebts...)

	if !finalRewards.IsZero() {
		lock.PosRewardDebts = truncatedTotalRewards
		lock.NegRewardDebts = sdk.NewCoins()
		k.SetLock(ctx, lock)

		err = k.bankKeeper.SendCoins(ctx, sdk.MustAccAddressFromBech32(key.PoolAddress), address, finalRewards)
		if err != nil {
			return nil, err
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeClaimRewards,
				sdk.NewAttribute(types.AttributeKeyLocker, msg.LockerAddress),
				sdk.NewAttribute(types.AttributeKeyKey, lock.Key),
				sdk.NewAttribute(sdk.AttributeKeyAmount, finalRewards.String()),
			),
		)
	}

	if !key.IsActive {
		k.DeleteLock(ctx, address, key.Name)

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

	address, err := sdk.AccAddressFromBech32(msg.LockerAddress)
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
