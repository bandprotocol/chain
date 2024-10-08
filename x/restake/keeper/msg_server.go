package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/restake/types"
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

	addr, err := sdk.AccAddressFromBech32(msg.StakerAddress)
	if err != nil {
		return nil, err
	}

	vault, found := k.GetVault(ctx, msg.Key)
	if !found {
		return nil, types.ErrVaultNotFound.Wrapf("key: %s", msg.Key)
	}

	lock, found := k.GetLock(ctx, addr, msg.Key)
	if !found {
		return nil, types.ErrLockNotFound.Wrapf("address: %s, key: %s", addr.String(), msg.Key)
	}

	reward := k.getReward(ctx, lock)
	finalRewards, remainders := reward.Rewards.TruncateDecimal()

	if !finalRewards.IsZero() {
		lock.PosRewardDebts = k.getAccumulatedRewards(ctx, lock)
		lock.NegRewardDebts = remainders
		k.SetLock(ctx, lock)

		err = k.bankKeeper.SendCoins(ctx, sdk.MustAccAddressFromBech32(vault.VaultAddress), addr, finalRewards)
		if err != nil {
			return nil, err
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeClaimRewards,
				sdk.NewAttribute(types.AttributeKeyStaker, msg.StakerAddress),
				sdk.NewAttribute(types.AttributeKeyKey, lock.Key),
				sdk.NewAttribute(types.AttributeKeyRewards, finalRewards.String()),
			),
		)
	}

	if !vault.IsActive {
		k.DeleteLock(ctx, addr, vault.Key)

		vault.Remainders = vault.Remainders.Add(remainders...)
		k.SetVault(ctx, vault)
	}

	return &types.MsgClaimRewardsResponse{}, nil
}
