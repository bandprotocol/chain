package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName defines the module name
	ModuleName = "tunnel"

	// Version defines the current version the IBC module supports
	Version = "bandchain-1"

	// KeyAccountsKey is used to store the key for the account
	KeyAccountsKey = "tunnel"

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

	TSSPacketCountStoreKey = append(GlobalStoreKeyPrefix, []byte("TSSPacketCount")...)

	AxelarPacketCountStoreKey = append(GlobalStoreKeyPrefix, []byte("AxelarPacketCount")...)

	TunnelStoreKeyPrefix = []byte{0x01}

	TSSPacketStoreKeyPrefix = []byte{0x02}

	AxelarPacketStoreKeyPrefix = []byte{0x03}

	ParamsKey = []byte{0x10}
)

func TunnelStoreKey(id uint64) []byte {
	return append(TunnelStoreKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func TSSPacketStoreKey(id uint64) []byte {
	return append(TSSPacketStoreKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func AxelarPacketStoreKey(id uint64) []byte {
	return append(AxelarPacketStoreKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}
