package keeper

import (
	minttypes "github.com/GeoDB-Limited/odin-core/x/mint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"
)

// Keeper of the mint store
type Keeper struct {
	cdc              codec.BinaryMarshaler
	storeKey         sdk.StoreKey
	paramSpace       paramtypes.Subspace
	stakingKeeper    minttypes.StakingKeeper
	authKeeper       minttypes.AccountKeeper
	bankKeeper       minttypes.BankKeeper
	feeCollectorName string
}

// NewKeeper creates a new mint Keeper instance
func NewKeeper(
	cdc codec.BinaryMarshaler, key sdk.StoreKey, paramSpace paramtypes.Subspace,
	sk minttypes.StakingKeeper, ak minttypes.AccountKeeper, bk minttypes.BankKeeper,
	feeCollectorName string,
) Keeper {
	// ensure mint module account is set
	if addr := ak.GetModuleAddress(minttypes.ModuleName); addr == nil {
		panic("the mint module account has not been set")
	}

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(minttypes.ParamKeyTable())
	}

	return Keeper{
		cdc:              cdc,
		storeKey:         key,
		paramSpace:       paramSpace,
		stakingKeeper:    sk,
		bankKeeper:       bk,
		authKeeper:       ak,
		feeCollectorName: feeCollectorName,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+minttypes.ModuleName)
}

// get the minter
func (k Keeper) GetMinter(ctx sdk.Context) (minter minttypes.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(minttypes.MinterKey)
	if b == nil {
		panic("stored minter should not have been nil")
	}

	k.cdc.MustUnmarshalBinaryBare(b, &minter)
	return
}

// set the minter
func (k Keeper) SetMinter(ctx sdk.Context, minter minttypes.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryBare(&minter)
	store.Set(minttypes.MinterKey, b)
}

// GetMintPool returns the mint pool info
func (k Keeper) GetMintPool(ctx sdk.Context) (mintPool minttypes.MintPool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(minttypes.MintPoolStoreKey)
	if b == nil {
		panic("Stored fee pool should not have been nil")
	}

	k.cdc.MustUnmarshalBinaryBare(b, &mintPool)
	return
}

// SetMintPool sets mint pool to the store
func (k Keeper) SetMintPool(ctx sdk.Context, mintPool minttypes.MintPool) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryBare(&mintPool)
	store.Set(minttypes.MintPoolStoreKey, b)
}

// GetParams returns the total set of minting parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params minttypes.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the total set of minting parameters.
func (k Keeper) SetParams(ctx sdk.Context, params minttypes.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetMintAccount returns the mint ModuleAccount
func (k Keeper) GetMintAccount(ctx sdk.Context) authtypes.ModuleAccountI {
	return k.authKeeper.GetModuleAccount(ctx, minttypes.ModuleName)
}

// SetMintAccount sets the module account
func (k Keeper) SetMintAccount(ctx sdk.Context, moduleAcc authtypes.ModuleAccountI) {
	k.authKeeper.SetModuleAccount(ctx, moduleAcc)
}

// StakingTokenSupply implements an alias call to the underlying staking keeper's
// StakingTokenSupply to be used in BeginBlocker.
func (k Keeper) StakingTokenSupply(ctx sdk.Context) sdk.Int {
	return k.stakingKeeper.StakingTokenSupply(ctx)
}

// BondedRatio implements an alias call to the underlying staking keeper's
// BondedRatio to be used in BeginBlocker.
func (k Keeper) BondedRatio(ctx sdk.Context) sdk.Dec {
	return k.stakingKeeper.BondedRatio(ctx)
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx sdk.Context, newCoins sdk.Coins) error {
	if newCoins.Empty() {
		// skip as no coins need to be minted
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, minttypes.ModuleName, newCoins)
}

// AddCollectedFees implements an alias call to the underlying supply keeper's
// AddCollectedFees to be used in BeginBlocker.
func (k Keeper) AddCollectedFees(ctx sdk.Context, fees sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, k.feeCollectorName, fees)
}

// LimitExceeded checks if withdrawal amount exceeds the limit
func (k Keeper) LimitExceeded(ctx sdk.Context, amt sdk.Coins) bool {
	moduleParams := k.GetParams(ctx)
	return amt.IsAnyGT(moduleParams.MaxWithdrawalPerTime)
}

// IsEligibleAccount checks if addr exists in the eligible to withdraw account pool
func (k Keeper) IsEligibleAccount(ctx sdk.Context, addr string) bool {
	params := k.GetParams(ctx)

	for _, item := range params.EligibleAccountsPool {
		if item == addr {
			return true
		}
	}

	return false
}

// WithdrawCoinsFromTreasury transfers coins from treasury pool to receiver account
func (k Keeper) WithdrawCoinsFromTreasury(ctx sdk.Context, receiver sdk.AccAddress, amount sdk.Coins) error {
	mintPool := k.GetMintPool(ctx)

	if amount.IsAllGT(mintPool.TreasuryPool) {
		return sdkerrors.Wrapf(
			minttypes.ErrWithdrawalAmountExceedsModuleBalance,
			"withdrawal amount: %s exceeds %s module balance",
			amount.String(),
			minttypes.ModuleName,
		)
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, receiver, amount); err != nil {
		return sdkerrors.Wrapf(
			err,
			"failed to withdraw %s from %s module account",
			amount.String(),
			minttypes.ModuleName,
		)
	}

	mintPool.TreasuryPool = mintPool.TreasuryPool.Sub(amount)
	k.SetMintPool(ctx, mintPool)

	return nil
}
