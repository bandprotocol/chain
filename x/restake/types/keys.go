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
)

var (
	GlobalStoreKeyPrefix = []byte{0x00}

	RemainderStoreKey = append(GlobalStoreKeyPrefix, []byte("Remainder")...)

	KeyStoreKeyPrefix    = []byte{0x01}
	StakeStoreKeyPrefix  = []byte{0x02}
	RewardStoreKeyPrefix = []byte{0x03}

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

func RewardsStoreKey(addr sdk.AccAddress) []byte {
	return append(RewardStoreKeyPrefix, address.MustLengthPrefix(addr)...)
}

func RewardStoreKey(addr sdk.AccAddress, keyName string) []byte {
	return append(RewardsStoreKey(addr), []byte(keyName)...)
}

func StakesByAmountIndexKey(addr sdk.AccAddress) []byte {
	return append(StakesByAmountIndexKeyPrefix, addr...)
}

func StakeByAmountIndexKey(stake Stake) []byte {
	address := sdk.MustAccAddressFromBech32(stake.Address)

	amountBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(amountBytes, stake.Amount.Uint64())

	// key is of format prefix || address || amountBytes || keyBytes
	bz := append(StakesByAmountIndexKey(address), amountBytes...)
	return append(bz, []byte(stake.Key)...)
}

func SplitRewardStoreKey(key []byte) ([]byte, sdk.AccAddress, string) {
	// <prefix (1 Byte)><addrLen (1 Byte)><addr_Bytes><key_Bytes>
	return key[0:1], sdk.AccAddress(key[2 : 2+key[1]]), string(key[2+key[1]:])
}
