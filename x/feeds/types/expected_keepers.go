package types

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
)

// AuthzKeeper defines the expected authz keeper. for query and testing only don't use to create/remove grant on deliver tx
type AuthzKeeper interface {
	GetAuthorization(
		ctx sdk.Context,
		feeder sdk.AccAddress,
		granter sdk.AccAddress,
		msgType string,
	) (authz.Authorization, *time.Time)
}

// OracleKeeper defines the expected oracle keeper
type OracleKeeper interface {
	MissReport(ctx sdk.Context, val sdk.ValAddress, requestTime time.Time)
	GetValidatorStatus(ctx sdk.Context, val sdk.ValAddress) oracletypes.ValidatorStatus
}

// StakingKeeper defines the expected staking keeper.
type StakingKeeper interface {
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool)
	IterateBondedValidatorsByPower(
		ctx sdk.Context,
		fn func(index int64, validator stakingtypes.ValidatorI) (stop bool),
	)
}

// RestakeKeeper defines the expected restake keeper.
type RestakeKeeper interface {
	SetLockedPower(ctx sdk.Context, addr sdk.AccAddress, key string, amount math.Int) error
}
