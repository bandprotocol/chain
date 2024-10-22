package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
)

// ValidatorInfo contains validator info and its index.
type ValidatorInfo struct {
	Index   int64
	Address sdk.ValAddress
	Power   uint64
	Status  oracletypes.ValidatorStatus
}

// NewValidatorInfo returns a new ValidatorInfo.
func NewValidatorInfo(
	index int64,
	address sdk.ValAddress,
	power uint64,
	status oracletypes.ValidatorStatus,
) ValidatorInfo {
	return ValidatorInfo{
		Index:   index,
		Address: address,
		Power:   power,
		Status:  status,
	}
}
