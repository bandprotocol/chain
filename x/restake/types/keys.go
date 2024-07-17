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

	KeyStoreKeyPrefix   = []byte{0x01}
	StakeStoreKeyPrefix = []byte{0x02}

	StakesByAmountIndexKeyPrefix = []byte{0x10}
)

func KeyStoreKey(keyName string) []byte {
	return append(KeyStoreKeyPrefix, []byte(keyName)...)
}

func StakesStoreKey(addr sdk.AccAddress) []byte {
	return append(StakeStoreKeyPrefix, address.MustLengthPrefix(addr)...)
}

func StakeStoreKey(addr sdk.AccAddress, keyName string) []byte {
	return append(StakesStoreKey(addr), []byte(keyName)...)
}

func StakesByAmountIndexKey(addr sdk.AccAddress) []byte {
	return append(StakesByAmountIndexKeyPrefix, address.MustLengthPrefix(addr)...)
}

func StakeByAmountIndexKey(stake Stake) []byte {
	address := sdk.MustAccAddressFromBech32(stake.Address)

	amountBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(amountBytes, stake.Amount.Uint64())

	// key is of format prefix || addrLen || address || amountBytes || keyBytes
	bz := append(StakesByAmountIndexKey(address), amountBytes...)
	return append(bz, []byte(stake.Key)...)
}
