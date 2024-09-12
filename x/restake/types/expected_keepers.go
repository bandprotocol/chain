package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	NewAccount(ctx sdk.Context, acc authtypes.AccountI) authtypes.AccountI
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	SetAccount(ctx sdk.Context, acc authtypes.AccountI)

	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) authtypes.ModuleAccountI
	SetModuleAccount(sdk.Context, authtypes.ModuleAccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}

// StakingKeeper defines the expected staking keeper.
type StakingKeeper interface {
	GetDelegatorBonded(ctx sdk.Context, delegator sdk.AccAddress) math.Int
	GetDelegation(
		ctx sdk.Context,
		delAddr sdk.AccAddress,
		valAddr sdk.ValAddress,
	) (delegation stakingtypes.Delegation, found bool)
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool)
}
