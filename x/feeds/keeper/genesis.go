package keeper

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	k.SetVotes(ctx, genState.Votes)

	signalTotalPowers := CalculateSignalTotalPowersFromVotes(genState.Votes)
	k.SetSignalTotalPowers(ctx, signalTotalPowers)

	feeds := k.CalculateNewCurrentFeeds(ctx)
	k.SetCurrentFeeds(ctx, feeds)

	if err := k.SetReferenceSourceConfig(ctx, genState.ReferenceSourceConfig); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the module's exported genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return types.NewGenesisState(
		k.GetParams(ctx),
		k.GetVotes(ctx),
		k.GetReferenceSourceConfig(ctx),
	)
}

// CalculateSignalTotalPowersFromVotes calculates the new signal-total-powers from the given votes using on init genesis process.
func CalculateSignalTotalPowersFromVotes(votes []types.Vote) []types.Signal {
	signalIDToPower := make(map[string]int64)
	for _, v := range votes {
		for _, signal := range v.Signals {
			signalIDToPower[signal.ID] += signal.Power
		}
	}

	keys := make([]string, 0, len(signalIDToPower))
	for k := range signalIDToPower {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	signalTotalPowers := []types.Signal{}
	for _, signalID := range keys {
		signalTotalPowers = append(signalTotalPowers, types.NewSignal(
			signalID,
			signalIDToPower[signalID],
		))
	}

	return signalTotalPowers
}
