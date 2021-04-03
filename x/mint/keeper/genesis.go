package keeper

import (
	"fmt"
	"github.com/GeoDB-Limited/odin-core/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(minter types.Minter, params types.Params, mintPool types.MintPool) *types.GenesisState {
	return &types.GenesisState{
		Minter:   minter,
		Params:   params,
		MintPool: mintPool,
	}
}

// InitGenesis new mint genesis
func InitGenesis(ctx sdk.Context, keeper Keeper, data *types.GenesisState) {
	keeper.SetMinter(ctx, data.Minter)
	keeper.SetParams(ctx, data.Params)

	moduleAcc := keeper.GetMintAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	balances := keeper.bankKeeper.GetAllBalances(ctx, moduleAcc.GetAddress())
	if balances.IsZero() {
		if err := keeper.bankKeeper.SetBalances(ctx, moduleAcc.GetAddress(), data.MintPool.TreasuryPool); err != nil {
			panic(err)
		}

		keeper.authKeeper.SetModuleAccount(ctx, moduleAcc)
	}

	keeper.SetMintPool(ctx, data.MintPool)
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() *types.GenesisState {
	return &types.GenesisState{
		Minter:   types.DefaultInitialMinter(),
		Params:   types.DefaultParams(),
		MintPool: types.InitialMintPool(),
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) *types.GenesisState {
	minter := keeper.GetMinter(ctx)
	params := keeper.GetParams(ctx)
	mintPool := keeper.GetMintPool(ctx)
	return NewGenesisState(minter, params, mintPool)
}

// ValidateGenesis validates the provided genesis state to ensure the
// expected invariants holds.
func ValidateGenesis(data types.GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	return types.ValidateMinter(data.Minter)
}
