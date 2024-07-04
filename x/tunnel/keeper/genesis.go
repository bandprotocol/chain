package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	k.SetTunnelCount(ctx, data.TunnelCount)
	// TODO: Set tunnels, axelarPacket and tssPacket

	k.SetTSSPacketCount(ctx, data.TssPacketCount)
	k.SetAxelarPacketCount(ctx, data.AxelarPacketCount)
}

// ExportGenesis returns the module's exported genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:            k.GetParams(ctx),
		TunnelCount:       k.GetTunnelCount(ctx),
		TssPacketCount:    k.GetTSSPacketCount(ctx),
		AxelarPacketCount: k.GetAxelarPacketCount(ctx),
	}
}
