package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v3/x/restake/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

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

	// check if all coins are allowed denom coins.
	allowedDenom := make(map[string]bool)
	for _, denom := range k.GetParams(ctx).AllowedDenoms {
		allowedDenom[denom] = true
	}

	for _, coin := range msg.Coins {
		if _, allow := allowedDenom[coin.Denom]; !allow {
			return nil, types.ErrNotAllowedDenom.Wrapf("denom: %s", coin.Denom)
		}
	}

	// transfer coins to module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, msg.Coins); err != nil {
		return nil, err
	}

	stake := k.GetStake(ctx, addr)
	stake.Coins = stake.Coins.Add(msg.Coins...)
	k.SetStake(ctx, stake)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeStake,
			sdk.NewAttribute(types.AttributeKeyStaker, addr.String()),
			sdk.NewAttribute(types.AttributeKeyCoins, msg.Coins.String()),
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

	// reduce staked coins. return error if unstake more than staked coins
	var isNeg bool
	stake := k.GetStake(ctx, addr)
	stake.Coins, isNeg = stake.Coins.SafeSub(msg.Coins...)
	if isNeg {
		return nil, types.ErrStakeNotEnough
	}

	if !stake.Coins.IsZero() {
		k.SetStake(ctx, stake)
	} else {
		k.DeleteStake(ctx, addr)
	}

	totalPower, err := k.GetTotalPower(ctx, addr)
	if err != nil {
		return nil, err
	}

	// check if total power is still more than locked power after unstaking.
	if !k.isValidPower(ctx, addr, totalPower) {
		return nil, types.ErrUnableToUnstake.Wrap("power is locked")
	}

	// transfer coins from module account to the account
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, msg.Coins); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnstake,
			sdk.NewAttribute(types.AttributeKeyStaker, addr.String()),
			sdk.NewAttribute(types.AttributeKeyCoins, msg.Coins.String()),
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
