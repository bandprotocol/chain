package bandtss

import (
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/bandtss/keeper"
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
)

// BeginBlocker handles the logic at the beginning of a block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	// Reward a portion of block rewards (inflation + tx fee) to active tss members.
	return k.AllocateTokens(ctx)
}

// EndBlocker handles tasks at the end of a block.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyEndBlocker)

	// execute group transition if the transition execution time is reached.
	if transition, ok := k.ShouldExecuteGroupTransition(ctx); ok {
		k.ExecuteGroupTransition(ctx, transition)
	}

	return nil
}
