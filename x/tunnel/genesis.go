package tunnel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// ValidateGenesis validates the provided genesis state.
func ValidateGenesis(data *types.GenesisState) error {
	return nil
}

func InitGenesis(ctx sdk.Context, k keeper.Keeper, data types.GenesisState) {
	err := k.SetParams(ctx, data.Params)
	if err != nil {
		panic(err)
	}
}

func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	return types.GenesisState{
		Params: k.GetParams(ctx),
	}
}
