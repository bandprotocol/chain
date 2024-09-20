package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// ValidateGenesis validates the provided genesis state.
func ValidateGenesis(data *types.GenesisState) error {
	if err := host.PortIdentifierValidator(data.PortID); err != nil {
		return err
	}

	// Validate the tunnel count
	if uint64(len(data.Tunnels)) != data.TunnelCount {
		return types.ErrInvalidGenesis.Wrapf("length of tunnels does not match tunnel count")
	}

	// validate the tunnel IDs
	for _, tunnel := range data.Tunnels {
		if tunnel.ID > data.TunnelCount {
			return types.ErrInvalidGenesis.Wrapf("tunnel count mismatch in tunnels")
		}
	}

	// validate the latest signal prices count
	if len(data.LatestSignalPricesList) != int(data.TunnelCount) {
		return types.ErrInvalidGenesis.Wrapf("length of latest signal prices does not match tunnel count")
	}

	// validate latest signal prices
	for _, latestSignalPrices := range data.LatestSignalPricesList {
		if latestSignalPrices.TunnelID == 0 || latestSignalPrices.TunnelID > data.TunnelCount {
			return types.ErrInvalidGenesis.Wrapf("tunnel count mismatch in latest signal prices")
		}
	}

	// validate the total fees
	if err := data.TotalFees.Validate(); err != nil {
		return types.ErrInvalidGenesis.Wrapf("invalid total fees: %s", err.Error())
	}

	return data.Params.Validate()
}

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k *Keeper, data *types.GenesisState) {
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
		if tunnel.IsActive {
			k.ActiveTunnelID(ctx, tunnel.ID)
		}
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

	// set the tunnels
	for _, tunnel := range data.Tunnels {
		k.ActiveTunnelID(ctx, tunnel.ID)
	}

	// set the latest signal prices
	for _, latestSignalPrices := range data.LatestSignalPricesList {
		k.SetLatestSignalPrices(ctx, latestSignalPrices)
	}

	// set the total fees
	k.SetTotalFees(ctx, data.TotalFees)
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k *Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:                 k.GetParams(ctx),
		PortID:                 types.PortID,
		TunnelCount:            k.GetTunnelCount(ctx),
		Tunnels:                k.GetTunnels(ctx),
		LatestSignalPricesList: k.GetAllLatestSignalPrices(ctx),
		TotalFees:              k.GetTotalFees(ctx),
	}
}
