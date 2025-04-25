package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/restake/types"
)

// Keeper of the x/restake store
type Keeper struct {
	storeKey      storetypes.StoreKey
	cdc           codec.BinaryCodec
	authKeeper    types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper

	authority string
}

// NewKeeper creates a new restake Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper,
	authority string,
) Keeper {
	// ensure module account is set
	if addr := authKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Errorf("invalid authority address: %w", err))
	}

	return Keeper{
		storeKey:      key,
		cdc:           cdc,
		authKeeper:    authKeeper,
		bankKeeper:    bankKeeper,
		stakingKeeper: stakingKeeper,
		authority:     authority,
	}
}

// GetModuleAccount returns the restake ModuleAccount
func (k Keeper) GetModuleAccount(ctx sdk.Context) sdk.ModuleAccountI {
	return k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
}

// SetModuleAccount sets a module account in the account keeper.
func (k Keeper) SetModuleAccount(ctx sdk.Context, acc sdk.ModuleAccountI) {
	k.authKeeper.SetModuleAccount(ctx, acc)
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetTotalPower returns the total power of an address.
func (k Keeper) GetTotalPower(ctx sdk.Context, stakerAddr sdk.AccAddress) (sdkmath.Int, error) {
	delegationPower, err := k.GetDelegationPower(ctx, stakerAddr)
	if err != nil {
		return sdkmath.Int{}, err
	}

	stakedPower := k.GetStakedPower(ctx, stakerAddr)
	return delegationPower.Add(stakedPower), nil
}

// GetDelegationPower returns the power from delegation
func (k Keeper) GetDelegationPower(ctx sdk.Context, stakerAddr sdk.AccAddress) (sdkmath.Int, error) {
	return k.stakingKeeper.GetDelegatorBonded(ctx, stakerAddr)
}

// IsLiquidStaker checks if an account associated with a given delegation is related to liquid staking
//
// This is determined by checking if the account has a 32-length address
// which will identify the following scenarios:
//   - An account has tokenized their shares, and thus the delegation is
//     owned by the tokenize share record module account
//   - A liquid staking provider is delegating through an ICA account
//
// Both ICA accounts and tokenize share record module accounts have 32-length addresses
func (k Keeper) IsLiquidStaker(addr sdk.AccAddress) bool {
	return len(addr) == 32
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}
