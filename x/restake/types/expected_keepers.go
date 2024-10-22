package types

import (
	"context"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
	SetModuleAccount(context.Context, sdk.ModuleAccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToAccount(
		ctx context.Context,
		senderModule string,
		recipientAddr sdk.AccAddress,
		amt sdk.Coins,
	) error
	SendCoinsFromAccountToModule(
		ctx context.Context,
		senderAddr sdk.AccAddress,
		recipientModule string,
		amt sdk.Coins,
	) error
}

// StakingKeeper defines the expected staking keeper.
type StakingKeeper interface {
	GetDelegatorBonded(ctx context.Context, delegator sdk.AccAddress) (math.Int, error)
	GetDelegation(
		ctx context.Context,
		delAddr sdk.AccAddress,
		valAddr sdk.ValAddress,
	) (stakingtypes.Delegation, error)
	GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error)
}
