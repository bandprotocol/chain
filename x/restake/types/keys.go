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

	// KeyAccountKey is the key used when generating a module address for the key
	KeyAccountsKey = "key-accounts"
)

var (
	GlobalStoreKeyPrefix = []byte{0x00}

	KeyStoreKeyPrefix  = []byte{0x01}
	LockStoreKeyPrefix = []byte{0x02}

	LocksByAmountIndexKeyPrefix = []byte{0x10}
)

func KeyStoreKey(keyName string) []byte {
	return append(KeyStoreKeyPrefix, []byte(keyName)...)
}

func LocksStoreKey(addr sdk.AccAddress) []byte {
	return append(LockStoreKeyPrefix, address.MustLengthPrefix(addr)...)
}

func LockStoreKey(addr sdk.AccAddress, keyName string) []byte {
	return append(LocksStoreKey(addr), []byte(keyName)...)
}

func LocksByAmountIndexKey(addr sdk.AccAddress) []byte {
	return append(LocksByAmountIndexKeyPrefix, address.MustLengthPrefix(addr)...)
}

func LockByAmountIndexKey(lock Lock) []byte {
	address := sdk.MustAccAddressFromBech32(lock.LockerAddress)

	amountBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(amountBytes, lock.Amount.Uint64())

	// key is of format prefix || addrLen || address || amountBytes || keyBytes
	bz := append(LocksByAmountIndexKey(address), amountBytes...)
	return append(bz, []byte(lock.Key)...)
}
