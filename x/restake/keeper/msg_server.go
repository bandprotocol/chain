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

	rewards := sdk.NewDecCoins()

	locks := k.GetLocks(ctx, address)
	for _, lock := range locks {
		key, err := k.GetKey(ctx, lock.Key)
		if err != nil {
			return nil, err
		}
		key = k.updateRewardPerShares(ctx, key)
		lock = k.updateRewardLefts(ctx, key, lock)
		rewards = rewards.Add(lock.RewardLefts...)
		lock.RewardLefts = sdk.NewDecCoins()
		k.SetLock(ctx, lock)
	}

	// truncate reward dec coins, return remainder to community pool
	finalRewards, remainder := rewards.TruncateDecimal()

	// add coins to user account
	if !finalRewards.IsZero() {
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, finalRewards)
		if err != nil {
			return nil, err
		}
	}
	k.addFeePool(ctx, remainder)

	return &types.MsgClaimRewardsResponse{}, nil
}
