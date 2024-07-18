package keeper

import (
	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

// Keeper of the x/restake store
type Keeper struct {
	storeKey      storetypes.StoreKey
	cdc           codec.BinaryCodec
	authKeeper    types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper
}

// NewKeeper creates a new restake Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper,
) Keeper {
	return Keeper{
		storeKey:      key,
		cdc:           cdc,
		authKeeper:    authKeeper,
		bankKeeper:    bankKeeper,
		stakingKeeper: stakingKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// SetLockedPower sets the new locked power amount of the address to the key
// This function will override the previous locked amount.
func (k Keeper) SetLockedPower(ctx sdk.Context, stakerAddr sdk.AccAddress, keyName string, amount math.Int) error {
	if !amount.IsUint64() {
		return types.ErrInvalidAmount
	}

	// check if delegation is not less than amount
	delegation := k.stakingKeeper.GetDelegatorBonded(ctx, stakerAddr)
	if delegation.LT(amount) {
		return types.ErrDelegationNotEnough
	}

	key, err := k.GetOrCreateKey(ctx, keyName)
	if err != nil {
		return err
	}

	if !key.IsActive {
		return types.ErrKeyNotActive
	}

	// check if there is a lock before
	stake, err := k.GetStake(ctx, stakerAddr, keyName)
	if err != nil {
		stake = types.Stake{
			StakerAddress:  stakerAddr.String(),
			Key:            keyName,
			Amount:         sdk.NewInt(0),
			PosRewardDebts: sdk.NewCoins(),
			NegRewardDebts: sdk.NewCoins(),
		}
	}

	key.TotalPower = key.TotalPower.Sub(stake.Amount).Add(amount)
	k.SetKey(ctx, key)

	diffLock := amount.Sub(stake.Amount)
	addtionalDebts := key.RewardPerPowers.MulDecTruncate(sdk.NewDecFromInt(diffLock.Abs()))
	truncatedAdditionalDebts, _ := addtionalDebts.TruncateDecimal()
	truncatedAdditionalDebts = truncatedAdditionalDebts.Sort()
	if diffLock.IsPositive() {
		stake.PosRewardDebts = stake.PosRewardDebts.Add(truncatedAdditionalDebts...)
	} else {
		stake.NegRewardDebts = stake.NegRewardDebts.Add(truncatedAdditionalDebts...)
	}
	stake.Amount = amount
	k.SetStake(ctx, stake)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLockPower,
			sdk.NewAttribute(types.AttributeKeyStaker, stakerAddr.String()),
			sdk.NewAttribute(types.AttributeKeyKey, keyName),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amount.String()),
		),
	)

	return nil
}

// AddRewards adds rewards to the pool address and re-calculate reward_per_share of the key
func (k Keeper) AddRewards(ctx sdk.Context, sender sdk.AccAddress, keyName string, rewards sdk.Coins) error {
	key, err := k.GetKey(ctx, keyName)
	if err != nil {
		return err
	}

	if !key.IsActive {
		return types.ErrKeyNotActive
	}

	if key.TotalPower.IsZero() {
		return types.ErrTotalLockZero
	}

	err = k.bankKeeper.SendCoins(ctx, sender, sdk.MustAccAddressFromBech32(key.PoolAddress), rewards)
	if err != nil {
		return err
	}

	key.RewardPerPowers = key.RewardPerPowers.Add(
		sdk.NewDecCoinsFromCoins(rewards.Sort()...).QuoDecTruncate(sdk.NewDecFromInt(key.TotalPower))...,
	)
	k.SetKey(ctx, key)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAddRewards,
			sdk.NewAttribute(types.AttributeKeyKey, keyName),
			sdk.NewAttribute(sdk.AttributeKeyAmount, rewards.String()),
		),
	)

	return nil
}

// GetLockedPower return locked power of the address to the key.
func (k Keeper) GetLockedPower(ctx sdk.Context, stakerAddr sdk.AccAddress, keyName string) (math.Int, error) {
	key, err := k.GetKey(ctx, keyName)
	if err != nil {
		return math.Int{}, types.ErrKeyNotFound
	}

	if !key.IsActive {
		return math.Int{}, types.ErrKeyNotActive
	}

	stake, err := k.GetStake(ctx, stakerAddr, keyName)
	if err != nil {
		return math.Int{}, types.ErrStakeNotFound
	}

	return stake.Amount, nil
}
