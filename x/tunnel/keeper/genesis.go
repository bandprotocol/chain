package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k Keeper, data *types.GenesisState) {
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	// check if the module account exists
	moduleAcc := k.GetTunnelAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}
	// set module account if its balance is zero
	if balance := k.GetModuleBalance(ctx); balance.IsZero() {
		k.SetModuleAccount(ctx, moduleAcc)
	} else if !balance.Equal(k.GetTotalDeposits(ctx).Add(k.GetTotalFees(ctx).TotalPacketFee...)) {
		panic("balance in the module account is not equal to the sum of total fees and total deposits")
	}

	// set the tunnel count
	k.SetTunnelCount(ctx, data.TunnelCount)

	// set tunnels
	for _, t := range data.Tunnels {
		k.SetTunnel(ctx, t)
		if t.IsActive {
			k.ActiveTunnelID(ctx, t.ID)
		}
	}

	// set the latest signal prices
	for _, latestSignalPrices := range data.LatestSignalPricesList {
		k.SetLatestSignalPrices(ctx, latestSignalPrices)
	}

	// set the deposits
	for _, deposit := range data.Deposits {
		k.SetDeposit(ctx, deposit)
	}

	// set the total fees
	k.SetTotalFees(ctx, data.TotalFees)
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:                 k.GetParams(ctx),
		TunnelCount:            k.GetTunnelCount(ctx),
		Tunnels:                k.GetTunnels(ctx),
		LatestSignalPricesList: k.GetAllLatestSignalPrices(ctx),
		Deposits:               k.GetAllDeposits(ctx),
		TotalFees:              k.GetTotalFees(ctx),
	}
}
