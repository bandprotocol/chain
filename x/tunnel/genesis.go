package tunnel

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"

	"github.com/bandprotocol/chain/v2/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// ValidateGenesis validates the provided genesis state.
func ValidateGenesis(data *types.GenesisState) error {
	if err := host.PortIdentifierValidator(data.PortID); err != nil {
		return err
	}

	// Validate the tunnel count
	if uint64(len(data.Tunnels)) != data.TunnelCount {
		return errorsmod.Wrapf(
			types.ErrInvalidGenesis,
			"TunnelCount: %d, actual tunnels: %d",
			data.TunnelCount,
			len(data.Tunnels),
		)
	}

	// Validate the tunnel IDs
	for _, tunnel := range data.Tunnels {
		if tunnel.ID > data.TunnelCount {
			return errorsmod.Wrapf(
				types.ErrInvalidGenesis,
				"TunnelID %d is greater than the TunnelCount %d",
				tunnel.ID,
				data.TunnelCount,
			)
		}
	}

	return data.Params.Validate()
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

	// Only try to bind to port if it is not already bound, since we may already own
	// port capability from capability InitGenesis
	if !k.HasCapability(ctx, types.PortID) {
		// tunnel module binds to the tunnel port on InitChain
		// and claims the returned capability
		err := k.BindPort(ctx, types.PortID)
		if err != nil {
			panic(fmt.Sprintf("could not claim port capability: %v", err))
		}
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		PortID:      types.PortID,
		Params:      k.GetParams(ctx),
		TunnelCount: k.GetTunnelCount(ctx),
		Tunnels:     k.GetTunnels(ctx),
	}
}
