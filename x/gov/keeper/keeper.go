package keeper

import (
	"fmt"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/tendermint/tendermint/libs/log"

	odingovtypes "github.com/GeoDB-Limited/odin-core/x/gov/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// Keeper defines the governance module Keeper
type Keeper struct {
	govkeeper.Keeper

	bankKeeper odingovtypes.BankKeeper
	mintKeeper odingovtypes.MintKeeper

	// The reference to the DelegationSet and ValidatorSet to get information about validators and delegators
	stakingKeeper odingovtypes.StakingKeeper

	// The (unexposed) keys used to access the stores from the Context.
	storeKey sdk.StoreKey

	// The codec codec for binary encoding/decoding.
	cdc codec.BinaryMarshaler
}

// NewKeeper returns a governance keeper. It handles:
// - submitting governance proposals
// - depositing funds into proposals, and activating upon sufficient funds being deposited
// - users voting on proposals, with weight proportional to stake in the system
// - and tallying the result of the vote.
//
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.BinaryMarshaler, key sdk.StoreKey, govKeeper govkeeper.Keeper,
	authKeeper govtypes.AccountKeeper, bankKeeper odingovtypes.BankKeeper,
	mintKeeper odingovtypes.MintKeeper, sk odingovtypes.StakingKeeper,
) Keeper {

	// ensure governance module account is set
	if addr := authKeeper.GetModuleAddress(govtypes.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", govtypes.ModuleName))
	}

	return Keeper{
		Keeper:        govKeeper,
		storeKey:      key,
		bankKeeper:    bankKeeper,
		mintKeeper:    mintKeeper,
		stakingKeeper: sk,
		cdc:           cdc,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+"odin"+govtypes.ModuleName)
}

// deleteVote deletes a vote from a given proposalID and voter from the store
func (k Keeper) deleteVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(govtypes.VoteKey(proposalID, voterAddr))
}
