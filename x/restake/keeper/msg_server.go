package keeper

import (
	"context"
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

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

// Stake defines a method for staking coins
func (k msgServer) Stake(
	goCtx context.Context,
	msg *types.MsgStake,
) (*types.MsgStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(msg.StakerAddress)
	if err != nil {
		return nil, err
	}

	coins := msg.Coins.Sort()

	// check if all coins are allowed denom coins.
	allowedDenoms := k.GetParams(ctx).AllowedDenoms
	for _, coin := range coins {
		if !slices.Contains(allowedDenoms, coin.Denom) {
			return nil, types.ErrNotAllowedDenom.Wrapf("expect: %s, got: %s", allowedDenoms, coin.Denom)
		}
	}

	// transfer coins to module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, coins); err != nil {
		return nil, err
	}

	stake := k.GetStake(ctx, addr)
	stake.Coins = stake.Coins.Add(coins...)
	k.SetStake(ctx, stake)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeStake,
			sdk.NewAttribute(types.AttributeKeyStaker, addr.String()),
			sdk.NewAttribute(types.AttributeKeyCoins, coins.String()),
		),
	)

	return &types.MsgStakeResponse{}, nil
}

// Unstake defines a method for unstaking coins
func (k msgServer) Unstake(
	goCtx context.Context,
	msg *types.MsgUnstake,
) (*types.MsgUnstakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(msg.StakerAddress)
	if err != nil {
		return nil, err
	}

	coins := msg.Coins.Sort()

	// reduce staked coins. return error if unstake more than staked coins
	var isNeg bool
	stake := k.GetStake(ctx, addr)
	stake.Coins, isNeg = stake.Coins.SafeSub(coins...)
	if isNeg {
		return nil, types.ErrStakeNotEnough
	}

	k.SetStake(ctx, stake)

	// check if total power is still more than locked power after unstaking.
	if !k.isValidPower(ctx, addr, k.GetTotalPower(ctx, addr)) {
		return nil, types.ErrUnableToUnstake.Wrap("power is locked")
	}

	// transfer coins from module account to the account
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, coins); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnstake,
			sdk.NewAttribute(types.AttributeKeyStaker, addr.String()),
			sdk.NewAttribute(types.AttributeKeyCoins, coins.String()),
		),
	)

	return &types.MsgUnstakeResponse{}, nil
}

// UpdateParams updates the module params.
func (k msgServer) UpdateParams(
	goCtx context.Context,
	req *types.MsgUpdateParams,
) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if k.GetAuthority() != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf(
			"invalid authority; expected %s, got %s",
			k.GetAuthority(),
			req.Authority,
		)
	}

	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
