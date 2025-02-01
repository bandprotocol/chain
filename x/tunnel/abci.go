package tunnel

import (
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyEndBlocker)

	// Produce packets for all tunnels that are active and have passed the interval time trigger
	// or deviated from the last price to destination route.
	// Error should not happen here since the tunnel is already validated.
	return k.ProduceActiveTunnelPackets(ctx)
}
