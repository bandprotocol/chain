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

	KeyStoreKeyPrefix  = []byte{0x01}
	LockStoreKeyPrefix = []byte{0x02}
)

func KeyStoreKey(keyName string) []byte {
	return append(KeyStoreKeyPrefix, []byte(keyName)...)
}

func LocksStoreKey(address sdk.AccAddress) []byte {
	return append(LockStoreKeyPrefix, address...)
}

func LockStoreKey(address sdk.AccAddress, keyName string) []byte {
	return append(LocksStoreKey(address), []byte(keyName)...)
}
