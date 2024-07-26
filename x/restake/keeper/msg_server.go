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

	reward := k.getReward(ctx, lock)
	finalRewards, remainders := reward.Rewards.TruncateDecimal()

	if !finalRewards.IsZero() {
		lock.PosRewardDebts = k.getAccumulatedRewards(ctx, lock)
		lock.NegRewardDebts = remainders
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
