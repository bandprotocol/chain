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

func (k Keeper) SetLockedToken(ctx sdk.Context, addr sdk.AccAddress, keyName string, amount math.Int) error {
	// check if delegation is not less than amount
	delegation := k.stakingKeeper.GetDelegatorBonded(ctx, addr)
	if delegation.LT(amount) {
		return types.ErrDelegationNotEnough
	}

	key := k.GetOrCreateKey(ctx, keyName)
	key = k.updateRewardPerShares(ctx, key)

	// check if there is a lock before
	// if yes, remove total lock in the key
	// if no, create the lock
	lock, err := k.GetLock(ctx, addr, keyName)
	if err != nil {
		lock = types.Lock{
			Address:     addr.String(),
			Key:         keyName,
			Amount:      amount,
			RewardDebts: key.RewardPerShares,
			RewardLefts: sdk.NewDecCoins(),
		}
		k.SetLock(ctx, lock)
	} else {
		lock = k.updateRewardLefts(ctx, key, lock)
		key.TotalLock = key.TotalLock.Sub(lock.Amount)
		lock.Amount = amount
		k.SetLock(ctx, lock)
	}

	// add total lock from new amount
	key.TotalLock = key.TotalLock.Add(lock.Amount)
	k.SetKey(ctx, key)

	return nil
}

func (k Keeper) AddRewards(ctx sdk.Context, keyName string, rewards sdk.Coins) error {
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, k.feeCollectorName, types.ModuleName, rewards)
	if err != nil {
		panic(err)
	}

	key, err := k.GetKey(ctx, keyName)
	if err != nil {
		return err
	}

	key.CurrentRewards = key.CurrentRewards.Add(sdk.NewDecCoinsFromCoins(rewards...)...)
	k.SetKey(ctx, key)

	return nil
}

func (k Keeper) GetLockedToken(ctx sdk.Context, addr sdk.AccAddress, keyName string) math.Int {
	lock, err := k.GetLock(ctx, addr, keyName)
	if err != nil {
		return sdk.NewInt(0)
	}

	return lock.Amount
}

func (k Keeper) CancelKey(ctx sdk.Context, keyName string) error {
	if !k.HasKey(ctx, keyName) {
		return types.ErrKeyNotFound
	}

	// TODO: can't delete as people won't be able to claim reward
	k.DeleteKey(ctx, keyName)

	return nil
}

func (k Keeper) addFeePool(ctx sdk.Context, decCoins sdk.DecCoins) {
	feePool := k.distrKeeper.GetFeePool(ctx)
	feePool.CommunityPool = feePool.CommunityPool.Add(decCoins...)
	k.distrKeeper.SetFeePool(ctx, feePool)
}
