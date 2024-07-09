package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	moduleAcc := k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	balance := k.bankKeeper.GetAllBalances(ctx, moduleAcc.GetAddress())
	if balance.IsZero() {
		k.authKeeper.SetModuleAccount(ctx, moduleAcc)
	}

	for _, key := range data.Keys {
		k.SetKey(ctx, key)
	}

	for _, stake := range data.Stakes {
		k.SetStake(ctx, stake)
	}

	for _, reward := range data.Rewards {
		address := sdk.MustAccAddressFromBech32(reward.Address)
		k.SetReward(ctx, address, types.Reward{
			Key:     reward.Key,
			Amounts: reward.Amounts,
		})
	}
}

// ExportGenesis returns the module's exported genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Keys:    k.GetKeys(ctx),
		Stakes:  k.GetAllStakes(ctx),
		Rewards: k.GetRewardsGenesis(ctx),
	}
}
