package tunnel

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// ValidateGenesis validates the provided genesis state.
func ValidateGenesis(data *types.GenesisState) error {
	// validate the tunnel count
	if uint64(len(data.Tunnels)) != data.TunnelCount {
		return types.ErrInvalidGenesis.Wrapf(
			"TunnelCount: %d, actual tunnels: %d",
			data.TunnelCount,
			len(data.Tunnels),
		)
	}

	// validate the tunnel IDs
	for _, tunnel := range data.Tunnels {
		if tunnel.ID > data.TunnelCount {
			return types.ErrInvalidGenesis.Wrapf(
				"TunnelID %d is greater than the TunnelCount %d",
				tunnel.ID,
				data.TunnelCount,
			)
		}
	}

	// validate the signal prices infos
	for _, signalPricesInfo := range data.SignalPricesInfos {
		if signalPricesInfo.TunnelID == 0 {
			return types.ErrInvalidGenesis.Wrapf(
				"TunnelID %d cannot be 0 or greater than the TunnelCount %d",
				signalPricesInfo.TunnelID,
				data.TunnelCount,
			)
		}
	}

	// validate the total fees
	err := data.TotalFees.TotalPacketFee.Validate()
	if err != nil {
		return err
	}

	return data.Params.Validate()
}

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k *keeper.Keeper, data *types.GenesisState) {
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
	}

	// set the tunnel count
	k.SetTunnelCount(ctx, data.TunnelCount)

	// set tunnels
	for _, tunnel := range data.Tunnels {
		k.SetTunnel(ctx, tunnel)
	}

	// set the tunnels
	for _, tunnel := range data.Tunnels {
		k.ActiveTunnelID(ctx, tunnel.ID)
	}

	// set the signal prices infos
	for _, signalPricesInfo := range data.SignalPricesInfos {
		k.SetSignalPricesInfo(ctx, signalPricesInfo)
	}

	// set the total fees
	k.SetTotalFees(ctx, data.TotalFees)
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:            k.GetParams(ctx),
		TunnelCount:       k.GetTunnelCount(ctx),
		Tunnels:           k.GetTunnels(ctx),
		SignalPricesInfos: k.GetSignalPricesInfos(ctx),
		TotalFees:         k.GetTotalFees(ctx),
	}
}
