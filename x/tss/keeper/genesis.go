package keeper

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// InitGenesis performs genesis initialization for the tss module.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	k.SetGroupGenesis(ctx, data.Groups)

	k.SetMembers(ctx, data.Members)

	desMapping := make(map[string][]types.DE)
	for _, deq := range data.DEs {
		desMapping[deq.Address] = append(desMapping[deq.Address], deq.DE)
	}

	// sort addresses to ensure consistent ordering
	addresses := make([]string, 0, len(desMapping))
	for addr := range desMapping {
		addresses = append(addresses, addr)
	}
	sort.Strings(addresses)

	for _, addr := range addresses {
		des := desMapping[addr]

		if uint64(len(des)) > data.Params.MaxDESize {
			panic(fmt.Sprintf("DEs of %s size exceeds MaxDESize", addr))
		}

		acc := sdk.MustAccAddressFromBech32(addr)
		deQueue := types.NewDEQueue(0, uint64(len(des)))
		k.SetDEQueue(ctx, acc, deQueue)

		for i, de := range des {
			k.SetDE(ctx, acc, uint64(i), de)
		}
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:  k.GetParams(ctx),
		Groups:  k.GetGroups(ctx),
		Members: k.GetMembers(ctx),
		DEs:     k.GetDEsGenesis(ctx),
	}
}

// GetDEsGenesis retrieves all DE from the context's KVStore.
func (k Keeper) GetDEsGenesis(ctx sdk.Context) []types.DEGenesis {
	var des []types.DEGenesis
	iterator := k.GetDEQueueIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		address := types.ExtractAddressFromDEQueueStoreKey(iterator.Key())
		var deQueue types.DEQueue
		k.cdc.MustUnmarshal(iterator.Value(), &deQueue)

		for i := deQueue.Head; i < deQueue.Tail; i++ {
			de, err := k.GetDE(ctx, address, i)
			if err != nil {
				panic(err)
			}
			des = append(des, types.DEGenesis{
				Address: address.String(),
				DE:      de,
			})
		}
	}
	return des
}

// SetGroupGenesis sets the group genesis state.
func (k Keeper) SetGroupGenesis(ctx sdk.Context, groups []types.Group) {
	maxGroupID := tss.GroupID(0)
	for _, g := range groups {
		if g.ID > maxGroupID {
			maxGroupID = g.ID
		}
		k.SetGroup(ctx, g)
	}
	k.SetGroupCount(ctx, uint64(maxGroupID))
}
