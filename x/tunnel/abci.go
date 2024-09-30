package tunnel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/keeper"
)

func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
}

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	// produce packets for all tunnels that are active and have passed the interval time trigger
	// or deviated from the last price to destination route
	k.ProduceActiveTunnelPackets(ctx)
}
