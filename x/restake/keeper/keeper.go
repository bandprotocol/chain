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

	key := k.GetOrCreateKey(ctx, keyName)
	if !key.IsActive {
		return types.ErrKeyNotActive
	}
	key = k.ProcessKey(ctx, key)

	// check if there is a lock before
	// if yes, update reward
	stake, err := k.GetStake(ctx, addr, keyName)
	if err == nil {
		k.ProcessStake(ctx, stake)
		key.TotalLock = key.TotalLock.Sub(stake.Amount)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeUnlockPower,
				sdk.NewAttribute(types.AttributeKeyAddress, addr.String()),
				sdk.NewAttribute(types.AttributeKeyKey, addr.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, stake.Amount.String()),
			),
		)
	}

	// update stake to new point
	stake = types.Stake{
		Address:     addr.String(),
		Key:         keyName,
		Amount:      amount,
		RewardDebts: key.RewardPerShares,
	}
	k.SetStake(ctx, stake)

	// add total lock from new amount
	key.TotalLock = key.TotalLock.Add(stake.Amount)
	k.SetKey(ctx, key)

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
	key := k.GetOrCreateKey(ctx, keyName)
	if !key.IsActive {
		return types.ErrKeyNotActive
	}

	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, rewards)
	if err != nil {
		return err
	}

	key.CurrentRewards = key.CurrentRewards.Add(sdk.NewDecCoinsFromCoins(rewards...)...)
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
