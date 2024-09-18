package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

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
	TunnelCountStoreKey = []byte{0x00}

	TotalPacketFeeStoreKey = []byte{0x01}

	ActiveTunnelIDStoreKeyPrefix = []byte{0x10}

	TunnelStoreKeyPrefix = []byte{0x11}

	PacketStoreKeyPrefix = []byte{0x12}

	LatestSignalPricesStoreKeyPrefix = []byte{0x13}

	DepositStoreKeyPrefix = []byte{0x14}

	ParamsKey = []byte{0x90}
)

func TunnelStoreKey(tunnelID uint64) []byte {
	return append(TunnelStoreKeyPrefix, sdk.Uint64ToBigEndian(tunnelID)...)
}

func ActiveTunnelIDStoreKey(tunnelID uint64) []byte {
	return append(ActiveTunnelIDStoreKeyPrefix, sdk.Uint64ToBigEndian(tunnelID)...)
}

func TunnelPacketsStoreKey(tunnelID uint64) []byte {
	return append(PacketStoreKeyPrefix, sdk.Uint64ToBigEndian(tunnelID)...)
}

func TunnelPacketStoreKey(tunnelID uint64, packetID uint64) []byte {
	return append(TunnelPacketsStoreKey(tunnelID), sdk.Uint64ToBigEndian(packetID)...)
}

func LatestSignalPricesStoreKey(tunnelID uint64) []byte {
	return append(LatestSignalPricesStoreKeyPrefix, sdk.Uint64ToBigEndian(tunnelID)...)
}

func DepositsStoreKey(tunnelID uint64) []byte {
	return append(DepositStoreKeyPrefix, sdk.Uint64ToBigEndian(tunnelID)...)
}

func DepositStoreKey(tunnelID uint64, depositor sdk.AccAddress) []byte {
	return append(DepositsStoreKey(tunnelID), address.MustLengthPrefix(depositor)...)
}
