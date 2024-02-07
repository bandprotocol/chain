package types

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
)

// OracleKeeper defines the expected oracle keeper
type OracleKeeper interface {
	MissReport(ctx sdk.Context, val sdk.ValAddress, requestTime time.Time)
	GetValidatorStatus(ctx sdk.Context, val sdk.ValAddress) oracletypes.ValidatorStatus
}

// StakingKeeper defines the expected staking keeper.
type StakingKeeper interface {
	GetBondedValidatorsByPower(ctx sdk.Context) []stakingtypes.Validator
	IterateBondedValidatorsByPower(
		ctx sdk.Context,
		fn func(index int64, validator stakingtypes.ValidatorI) (stop bool),
	)
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool)
	GetDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress, maxRetrieve uint16) []stakingtypes.Delegation
	GetDelegatorBonded(ctx sdk.Context, delegator sdk.AccAddress) math.Int
}
