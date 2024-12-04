package keeper

import (
	"fmt"
	"strings"

	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// bindIBCPort will reserve the port.
// returns a string name of the port or error if we cannot bind it.
// this will fail if call twice.
func (k Keeper) bindIBCPort(ctx sdk.Context, portID string) error {
	portCap := k.portKeeper.BindPort(ctx, portID)
	return k.ClaimCapability(ctx, portCap, host.PortPath(portID))
}

// ensureIBCPort is like bindIBCPort, but it checks if we already hold the port
// before calling register, so this is safe to call multiple times.
// Returns success if we already registered or just registered and error if we cannot
// (lack of permissions or someone else has it)
func (k Keeper) ensureIBCPort(ctx sdk.Context, tunnelID uint64) (string, error) {
	portID := PortIDForTunnel(tunnelID)
	if _, ok := k.scopedKeeper.GetCapability(ctx, host.PortPath(portID)); ok {
		return portID, nil
	}
	return portID, k.bindIBCPort(ctx, portID)
}

const portIDPrefix = "tunnel."

// PortIDForTunnel generates a unique port ID for a given tunnel ID.
// It concatenates a predefined prefix (portIDPrefix) with the tunnel ID.
func PortIDForTunnel(tunnelID uint64) string {
	return fmt.Sprintf("%s%d", portIDPrefix, tunnelID)
}

// IsValidPortID checks if a given port ID is valid.
// It ensures that the port ID starts with the predefined prefix (portIDPrefix).
func IsValidPortID(portID string) bool {
	return strings.HasPrefix(portID, portIDPrefix)
}

// AuthenticateCapability wraps the scopedKeeper's AuthenticateCapability function
func (k Keeper) AuthenticateCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) bool {
	return k.scopedKeeper.AuthenticateCapability(ctx, cap, name)
}

// ClaimCapability allows the tunnel module that can claim a capability that IBC module
// passes to it
func (k Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}
