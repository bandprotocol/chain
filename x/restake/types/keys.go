package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
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

	// VaultAccountsKey is the key used when generating a module address for the vault
	VaultAccountsKey = "vault-accounts"
)

var (
	GlobalStoreKeyPrefix = []byte{0x00}

	VaultStoreKeyPrefix = []byte{0x01}
	LockStoreKeyPrefix  = []byte{0x02}

	LocksByPowerIndexKeyPrefix = []byte{0x10}
)

// VaultStoreKey returns the key to retrieve a specific vault from the store.
func VaultStoreKey(key string) []byte {
	return append(VaultStoreKeyPrefix, []byte(key)...)
}

// LocksByAddressStoreKey returns the key to retrieve all locks of an address from the store.
func LocksByAddressStoreKey(addr sdk.AccAddress) []byte {
	return append(LockStoreKeyPrefix, address.MustLengthPrefix(addr)...)
}

// LockStoreKey returns the key to retrieve a lock of an address and the key from the store.
func LockStoreKey(addr sdk.AccAddress, key string) []byte {
	return append(LocksByAddressStoreKey(addr), []byte(key)...)
}

// LocksByPowerIndexKey returns the key to retrieve all locks of an address ordering by locked power from the store.
func LocksByPowerIndexKey(addr sdk.AccAddress) []byte {
	return append(LocksByPowerIndexKeyPrefix, address.MustLengthPrefix(addr)...)
}

// LockByPowerIndexKey returns the key to retrieve a lock by power from the store.
func LockByPowerIndexKey(lock Lock) []byte {
	address := sdk.MustAccAddressFromBech32(lock.StakerAddress)

	powerBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(powerBytes, lock.Power.Uint64())

	// key is of format prefix || addrLen || address || powerBytes || keyBytes
	bz := append(LocksByPowerIndexKey(address), powerBytes...)
	return append(bz, []byte(lock.Key)...)
}
