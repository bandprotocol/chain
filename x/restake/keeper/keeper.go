package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

// Keeper of the x/restake store
type Keeper struct {
	storeKey         storetypes.StoreKey
	cdc              codec.BinaryCodec
	feeCollectorName string
	authKeeper       types.AccountKeeper
	bankKeeper       types.BankKeeper
	distrKeeper      types.DistrKeeper
	stakingKeeper    types.StakingKeeper
	authority        string
}

// NewKeeper creates a new restake Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	feeCollectorName string,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	dk types.DistrKeeper,
	sk types.StakingKeeper,
	authority string,
) Keeper {
	// ensure module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// ensure that authority is a valid AccAddress
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic("authority is not a valid acc address")
	}

	return Keeper{
		storeKey:         key,
		cdc:              cdc,
		feeCollectorName: feeCollectorName,
		authKeeper:       ak,
		bankKeeper:       bk,
		distrKeeper:      dk,
		stakingKeeper:    sk,
		authority:        authority,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) SetLockedPower(ctx sdk.Context, addr sdk.AccAddress, keyName string, amount math.Int) error {
	if !amount.IsUint64() {
		return types.ErrInvalidAmount
	}

	// check if delegation is not less than amount
	delegation := k.stakingKeeper.GetDelegatorBonded(ctx, addr)
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
	stake, err := k.GetStake(ctx, addr, keyName)
	if err != nil {
		stake = types.Stake{
			Address:        addr.String(),
			Key:            keyName,
			Amount:         sdk.NewInt(0),
			PosRewardDebts: sdk.NewCoins(),
			NegRewardDebts: sdk.NewCoins(),
		}
	}

	key.TotalLock = key.TotalLock.Sub(stake.Amount).Add(amount)
	k.SetKey(ctx, key)

	diffLock := amount.Sub(stake.Amount)
	addtionalDebts := key.RewardPerShares.MulDecTruncate(sdk.NewDecFromInt(diffLock.Abs()))
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
			sdk.NewAttribute(types.AttributeKeyAddress, addr.String()),
			sdk.NewAttribute(types.AttributeKeyKey, addr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amount.String()),
		),
	)

	return nil
}

func (k Keeper) AddRewards(ctx sdk.Context, sender sdk.AccAddress, keyName string, rewards sdk.Coins) error {
	key, err := k.GetKey(ctx, keyName)
	if err != nil {
		return err
	}

	if !key.IsActive {
		return types.ErrKeyNotActive
	}

	if key.TotalLock.IsZero() {
		return types.ErrTotalLockZero
	}

	err = k.bankKeeper.SendCoins(ctx, sender, sdk.MustAccAddressFromBech32(key.Address), rewards)
	if err != nil {
		return err
	}

	key.RewardPerShares = key.RewardPerShares.Add(
		sdk.NewDecCoinsFromCoins(rewards.Sort()...).QuoDecTruncate(sdk.NewDecFromInt(key.TotalLock))...,
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

func (k Keeper) GetLockedPower(ctx sdk.Context, addr sdk.AccAddress, keyName string) (math.Int, error) {
	key, err := k.GetKey(ctx, keyName)
	if err != nil {
		return math.Int{}, types.ErrKeyNotFound
	}

	if !key.IsActive {
		return math.Int{}, types.ErrKeyNotActive
	}

	stake, err := k.GetStake(ctx, addr, keyName)
	if err != nil {
		return math.Int{}, types.ErrStakeNotFound
	}

	return stake.Amount, nil
}
