package types

import (
	minttypes "github.com/GeoDB-Limited/odin-core/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankexported "github.com/cosmos/cosmos-sdk/x/bank/exported"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// StakingKeeper expected staking keeper (Validator and Delegator sets) (noalias)
type StakingKeeper interface {
	// IterateBondedValidatorsByPower iterate through bonded validators by operator address, execute func for each validator
	IterateBondedValidatorsByPower(
		sdk.Context, func(index int64, validator stakingtypes.ValidatorI) (stop bool),
	)

	TotalBondedTokens(sdk.Context) sdk.Int // total bonded tokens within the validator set
	IterateDelegations(
		ctx sdk.Context, delegator sdk.AccAddress,
		fn func(index int64, delegation stakingtypes.DelegationI) (stop bool),
	)
	BondDenom(ctx sdk.Context) string
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetSupply(ctx sdk.Context) bankexported.SupplyI
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

type MintKeeper interface {
	GetMintPool(ctx sdk.Context) minttypes.MintPool
}
