package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName defines the module name
	ModuleName = "tunnel"

	// Version defines the current version the IBC module supports
	Version = "tunnel-1"

	// TunnelAccountsKey is used to store the key for the account
	TunnelAccountsKey = "tunnel-accounts"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the tunnel module
	QuerierRoute = ModuleName

	// PortID is the default port id that oracle module binds to.
	PortID = ModuleName
)

var (
	// global store keys
	TunnelCountStoreKey    = []byte{0x00}
	TotalPacketFeeStoreKey = []byte{0x01}

	// store prefixes
	ActiveTunnelIDStoreKeyPrefix     = []byte{0x10}
	TunnelStoreKeyPrefix             = []byte{0x11}
	PacketStoreKeyPrefix             = []byte{0x12}
	LatestSignalPricesStoreKeyPrefix = []byte{0x13}

	// params store keys
	ParamsKey = []byte{0x90}
)

// TunnelStoreKey returns the key to retrieve a specific tunnel from the store.
func TunnelStoreKey(tunnelID uint64) []byte {
	return append(TunnelStoreKeyPrefix, sdk.Uint64ToBigEndian(tunnelID)...)
}

// ActiveTunnelIDStoreKey returns the key to retrieve a specific active tunnel ID from the store.
func ActiveTunnelIDStoreKey(tunnelID uint64) []byte {
	return append(ActiveTunnelIDStoreKeyPrefix, sdk.Uint64ToBigEndian(tunnelID)...)
}

// TunnelPacketsStoreKey returns the key to retrieve all packets of a tunnel from the store.
func TunnelPacketsStoreKey(tunnelID uint64) []byte {
	return append(PacketStoreKeyPrefix, sdk.Uint64ToBigEndian(tunnelID)...)
}

// TunnelPacketStoreKey returns the key to retrieve a packet of a tunnel from the store.
func TunnelPacketStoreKey(tunnelID uint64, packetID uint64) []byte {
	return append(TunnelPacketsStoreKey(tunnelID), sdk.Uint64ToBigEndian(packetID)...)
}

// LatestSignalPricesStoreKey returns the key to retrieve the latest signal prices from the store.
func LatestSignalPricesStoreKey(tunnelID uint64) []byte {
	return append(LatestSignalPricesStoreKeyPrefix, sdk.Uint64ToBigEndian(tunnelID)...)
}
