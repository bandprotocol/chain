package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "restake"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the restake module
	QuerierRoute = ModuleName
)

var (
	GlobalStoreKeyPrefix = []byte{0x00}

	RemainderStoreKey = append(GlobalStoreKeyPrefix, []byte("Remainder")...)

	KeyStoreKeyPrefix   = []byte{0x01}
	StakeStoreKeyPrefix = []byte{0x02}
)

func KeyStoreKey(keyName string) []byte {
	return append(KeyStoreKeyPrefix, []byte(keyName)...)
}

func StakesStoreKey(address sdk.AccAddress) []byte {
	return append(StakeStoreKeyPrefix, address...)
}

func StakeStoreKey(address sdk.AccAddress, keyName string) []byte {
	return append(StakesStoreKey(address), []byte(keyName)...)
}
