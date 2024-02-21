package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI

	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) types.ModuleAccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin

	SendCoinsFromModuleToAccount(
		ctx sdk.Context,
		senderModule string,
		recipientAddr sdk.AccAddress,
		amt sdk.Coins,
	) error
	SendCoinsFromAccountToModule(
		ctx sdk.Context,
		senderAddr sdk.AccAddress,
		recipientModule string,
		amt sdk.Coins,
	) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
}

// DistrKeeper defines the expected distribution keeper.
type DistrKeeper interface {
	GetCommunityTax(ctx sdk.Context) (percent sdk.Dec)
	GetFeePool(ctx sdk.Context) (feePool distrtypes.FeePool)
	SetFeePool(ctx sdk.Context, feePool distrtypes.FeePool)
	AllocateTokensToValidator(ctx sdk.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins)
}

// RollingseedKeeper defines the expected rollingseed keeper
type RollingseedKeeper interface {
	GetRollingSeed(ctx sdk.Context) []byte
}

// StakingKeeper defines the expected staking keeper.
type StakingKeeper interface {
	MaxValidators(ctx sdk.Context) (res uint32)
	ValidatorByConsAddr(sdk.Context, sdk.ConsAddress) stakingtypes.ValidatorI
	IterateBondedValidatorsByPower(
		ctx sdk.Context,
		fn func(index int64, validator stakingtypes.ValidatorI) (stop bool),
	)
}

// TSSKeeper defines the expected tss keeper (noalias)
type TSSKeeper interface {
	CreateGroup(ctx sdk.Context, input tsstypes.CreateGroupInput) (*tsstypes.CreateGroupResult, error)
	CreateSigning(ctx sdk.Context, input tsstypes.CreateSigningInput) (*tsstypes.CreateSigningResult, error)
	UpdateGroupFee(ctx sdk.Context, input tsstypes.UpdateGroupFeeInput) (*tsstypes.UpdateGroupFeeResult, error)
	ReplaceGroup(ctx sdk.Context, input tsstypes.ReplaceGroupInput) (*tsstypes.ReplaceGroupResult, error)

	GetSigningCount(ctx sdk.Context) uint64
	GetStatus(ctx sdk.Context, address sdk.AccAddress) tsstypes.Status
	GetActiveGroup(ctx sdk.Context, groupID tss.GroupID) (tsstypes.Group, error)
	GetReplacement(ctx sdk.Context, replacementID uint64) (tsstypes.Replacement, error)
	GetActiveMembers(ctx sdk.Context, groupID tss.GroupID) ([]tsstypes.Member, error)

	SetActiveStatus(ctx sdk.Context, address sdk.AccAddress) error
	SetInactiveStatus(ctx sdk.Context, address sdk.AccAddress)
	SetLastActive(ctx sdk.Context, address sdk.AccAddress) error
}
