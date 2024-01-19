package types

import (
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
)

// OracleKeeper defines the expected oracle keeper
type OracleKeeper interface {
	MissReport(ctx sdk.Context, val sdk.ValAddress, requestTime time.Time)
	GetValidatorStatus(ctx sdk.Context, val sdk.ValAddress) oracletypes.ValidatorStatus
}
