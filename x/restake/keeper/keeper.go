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
func (k Keeper) SetLockedPower(ctx sdk.Context, lockerAddr sdk.AccAddress, keyName string, amount math.Int) error {
	if !amount.IsUint64() {
		return types.ErrInvalidAmount
	}

	// check if delegation is not less than amount
	delegation := k.stakingKeeper.GetDelegatorBonded(ctx, lockerAddr)
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
	lock, err := k.GetLock(ctx, lockerAddr, keyName)
	if err != nil {
		lock = types.Lock{
			LockerAddress:  lockerAddr.String(),
			Key:            keyName,
			Amount:         sdk.NewInt(0),
			PosRewardDebts: sdk.NewCoins(),
			NegRewardDebts: sdk.NewCoins(),
		}
	}

	key.TotalPower = key.TotalPower.Sub(lock.Amount).Add(amount)
	k.SetKey(ctx, key)

	diffAmount := amount.Sub(lock.Amount)
	addtionalDebts := key.RewardPerPowers.MulDecTruncate(sdk.NewDecFromInt(diffAmount.Abs()))
	truncatedAdditionalDebts, _ := addtionalDebts.TruncateDecimal()
	truncatedAdditionalDebts = truncatedAdditionalDebts.Sort()
	if diffAmount.IsPositive() {
		lock.PosRewardDebts = lock.PosRewardDebts.Add(truncatedAdditionalDebts...)
	} else {
		lock.NegRewardDebts = lock.NegRewardDebts.Add(truncatedAdditionalDebts...)
	}
	lock.Amount = amount
	k.SetLock(ctx, lock)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLockPower,
			sdk.NewAttribute(types.AttributeKeyLocker, lockerAddr.String()),
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

	decRewards := sdk.NewDecCoinsFromCoins(rewards.Sort()...)
	totalPower := sdk.NewDecFromInt(key.TotalPower)
	rewardPerPowers := decRewards.QuoDecTruncate(totalPower)
	truncatedRewards := decRewards.Sub(rewardPerPowers.MulDecTruncate(totalPower))

	// add truncate part to remainder
	// e.g. rewards = 1, totalPower = 3 -> rewardPerPower = 0.333333333333333
	// remainder = 1 - (0.333333333333333 * 3)
	key.Remainders = key.Remainders.Add(truncatedRewards...)
	key.RewardPerPowers = key.RewardPerPowers.Add(rewardPerPowers...)
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

// GetLockedPower returns locked power of the address to the key.
func (k Keeper) GetLockedPower(ctx sdk.Context, lockerAddr sdk.AccAddress, keyName string) (math.Int, error) {
	key, err := k.GetKey(ctx, keyName)
	if err != nil {
		return math.Int{}, types.ErrKeyNotFound
	}

	if !key.IsActive {
		return math.Int{}, types.ErrKeyNotActive
	}

	lock, err := k.GetLock(ctx, lockerAddr, keyName)
	if err != nil {
		return math.Int{}, types.ErrLockNotFound
	}

	return lock.Amount, nil
}
