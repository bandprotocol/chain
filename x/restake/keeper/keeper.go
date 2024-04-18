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
	authority     string
}

// NewKeeper creates a new restake Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	sk types.StakingKeeper,
	authority string,
) Keeper {
	// ensure that authority is a valid AccAddress
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic("authority is not a valid acc address")
	}

	return Keeper{
		storeKey:      key,
		cdc:           cdc,
		authKeeper:    ak,
		bankKeeper:    bk,
		stakingKeeper: sk,
		authority:     authority,
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

	// create key if it doesn't exist
	if !k.HasKey(ctx, keyName) {
		k.SetKey(ctx, types.Key{
			Name:     keyName,
			IsActive: true,
		})
	}

	k.SetLock(ctx, types.Lock{
		Address: addr.String(),
		Key:     keyName,
		Amount:  amount,
	})

	return nil
}

func (k Keeper) DistributeToken(ctx sdk.Context, keyName string, amount sdk.Coins) error {
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

	k.DeleteKey(ctx, keyName)

	return nil
}
