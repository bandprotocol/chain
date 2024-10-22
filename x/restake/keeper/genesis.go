package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/restake/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	// check if the module account exists
	moduleAcc := k.GetModuleAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	balances := k.bankKeeper.GetAllBalances(ctx, moduleAcc.GetAddress())
	if balances.IsZero() {
		k.authKeeper.SetModuleAccount(ctx, moduleAcc)
	}

	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	for _, vault := range data.Vaults {
		k.SetVault(ctx, vault)
	}

	for _, lock := range data.Locks {
		k.SetLock(ctx, lock)
	}

	var totalStakes sdk.Coins
	for _, stake := range data.Stakes {
		k.SetStake(ctx, stake)
		totalStakes = totalStakes.Add(stake.Coins...)
	}

	// check if total stakes equals balances of the module account.
	if !balances.Equal(totalStakes) {
		panic(fmt.Sprintf("expected module account was %s but we got %s", balances.String(), totalStakes.String()))
	}
}

// ExportGenesis returns the module's exported genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return types.NewGenesisState(
		k.GetParams(ctx),
		k.GetVaults(ctx),
		k.GetLocks(ctx),
		k.GetStakes(ctx),
	)
}
