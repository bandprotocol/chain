package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
)

// ValidatorInfo contains validator info and its index.
type ValidatorInfo struct {
	Index   int64
	Address sdk.ValAddress
	Power   uint64
	Status  oracletypes.ValidatorStatus
}
