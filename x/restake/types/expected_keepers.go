package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// DistributionKeeper expected distribution keeper (noalias)
type DistributionKeeper interface {
	GetFeePoolCommunityCoins(ctx sdk.Context) sdk.DecCoins
	GetValidatorOutstandingRewardsCoins(ctx sdk.Context, val sdk.ValAddress) sdk.DecCoins
}

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	IterateAccounts(ctx sdk.Context, process func(authtypes.AccountI) (stop bool))
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI // only used for simulation

	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI

	SetModuleAccount(sdk.Context, authtypes.ModuleAccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SendCoinsFromModuleToModule(ctx sdk.Context, senderPool, recipientPool string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(
		ctx sdk.Context,
		senderModule string,
		recipientAddr sdk.AccAddress,
		amt sdk.Coins,
	) error
}

// StakingKeeper defines the expected staking keeper.
type StakingKeeper interface {
	GetBondedValidatorsByPower(ctx sdk.Context) []stakingtypes.Validator
	IterateBondedValidatorsByPower(
		ctx sdk.Context,
		fn func(index int64, validator stakingtypes.ValidatorI) (stop bool),
	)
	GetDelegatorBonded(ctx sdk.Context, delegator sdk.AccAddress) math.Int
}

// DistrKeeper defines the expected distribution keeper.
type DistrKeeper interface {
	GetCommunityTax(ctx sdk.Context) (percent sdk.Dec)
	GetFeePool(ctx sdk.Context) (feePool distrtypes.FeePool)
	SetFeePool(ctx sdk.Context, feePool distrtypes.FeePool)
	AllocateTokensToValidator(ctx sdk.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins)
}
