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
	GlobalStoreKeyPrefix = []byte{0x00}

	TunnelCountStoreKey = append(GlobalStoreKeyPrefix, []byte("TunnelCount")...)

	PendingTriggerTunnelsStoreKey = append(GlobalStoreKeyPrefix, []byte("PendingTriggerTunnels")...)

	TunnelStoreKeyPrefix = []byte{0x01}

	PacketStoreKeyPrefix = []byte{0x02}

	ParamsKey = []byte{0x10}
)

func TunnelStoreKey(tunnelID uint64) []byte {
	return append(TunnelStoreKeyPrefix, sdk.Uint64ToBigEndian(tunnelID)...)
}

func TunnelPacketsStoreKey(tunnelID uint64) []byte {
	return append(PacketStoreKeyPrefix, sdk.Uint64ToBigEndian(tunnelID)...)
}

func TunnelPacketStoreKey(tunnelID uint64, packetID uint64) []byte {
	return append(TunnelPacketsStoreKey(tunnelID), sdk.Uint64ToBigEndian(packetID)...)
}
