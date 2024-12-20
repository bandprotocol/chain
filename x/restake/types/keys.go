package types

import (
	"encoding/binary"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/kv"
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
	VaultStoreKeyPrefix = []byte{0x10}
	LockStoreKeyPrefix  = []byte{0x11}
	StakeStoreKeyPrefix = []byte{0x12}

	LocksByPowerIndexKeyPrefix = []byte{0x80}

	ParamsKey = []byte{0x90}
)

// VaultStoreKey returns the key to retrieve a specified vault from the store.
func VaultStoreKey(key string) []byte {
	return append(VaultStoreKeyPrefix, []byte(key)...)
}

// StakeStoreKey returns the key to retrieve the stake of an address from the store.
func StakeStoreKey(addr sdk.AccAddress) []byte {
	return append(StakeStoreKeyPrefix, address.MustLengthPrefix(addr)...)
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

	// the format of key is prefix || addrLen || address || powerBytes || keyBytes
	bz := append(LocksByPowerIndexKey(address), powerBytes...)
	return append(bz, []byte(lock.Key)...)
}

// SplitLockByPowerIndexKey split the LockByPowerIndexKey and returns the address and power
func SplitLockByPowerIndexKey(key []byte) (addr sdk.AccAddress, power sdkmath.Int) {
	// the format of key is prefix || addrLen || address || powerBytes || keyBytes
	kv.AssertKeyAtLeastLength(key, 2)
	addrLen := int(key[1])

	kv.AssertKeyAtLeastLength(key, 2+addrLen+8)
	addr = sdk.AccAddress(key[2 : 2+addrLen])
	power = sdkmath.NewIntFromUint64(binary.BigEndian.Uint64(key[2+addrLen : 2+addrLen+8]))

	return
}
