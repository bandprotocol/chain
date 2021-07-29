package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (r *QueryExtendedValidatorsResponse) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return stakingtypes.Validators(r.Validators).UnpackInterfaces(unpacker)
}
