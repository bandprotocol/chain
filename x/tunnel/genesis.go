package tunnel

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// ValidateGenesis validates the provided genesis state.
func ValidateGenesis(data *types.GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return errorsmod.Wrap(err, "tunnel params")
	}
	return nil
}

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	k.SetTunnelCount(ctx, data.TunnelCount)

	for _, tunnel := range data.Tunnels {
		k.SetTunnel(ctx, tunnel)
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:      k.GetParams(ctx),
		TunnelCount: k.GetTunnelCount(ctx),
		Tunnels:     k.GetTunnels(ctx),
	}
}
