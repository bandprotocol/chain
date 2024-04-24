package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "feeds"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the feeds module
	QuerierRoute = ModuleName
)

var (
	GlobalStoreKeyPrefix = []byte{0x00}

	PriceServiceStoreKey = append(GlobalStoreKeyPrefix, []byte("PriceService")...)

	FeedStoreKeyPrefix            = []byte{0x01}
	PriceValidatorStoreKeyPrefix  = []byte{0x02}
	PriceStoreKeyPrefix           = []byte{0x03}
	DelegatorSignalStoreKeyPrefix = []byte{0x04}

	ParamsKey = []byte{0x10}

	FeedsByPowerIndexKeyPrefix = []byte{0x20}
)

func DelegatorSignalStoreKey(delegator sdk.AccAddress) []byte {
	return append(DelegatorSignalStoreKeyPrefix, delegator...)
}

func FeedStoreKey(signalID string) []byte {
	return append(FeedStoreKeyPrefix, []byte(signalID)...)
}

func PriceValidatorsStoreKey(signalID string) []byte {
	return append(PriceValidatorStoreKeyPrefix, []byte(signalID)...)
}

func PriceValidatorStoreKey(signalID string, validator sdk.ValAddress) []byte {
	return append(PriceValidatorsStoreKey(signalID), validator...)
}

func PriceStoreKey(signalID string) []byte {
	return append(PriceStoreKeyPrefix, []byte(signalID)...)
}

func FeedsByPowerIndexKey(signalID string, power uint64) []byte {
	powerBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(powerBytes, power)
	powerBytesLen := len(powerBytes) // 8

	signalIDBytes := []byte(signalID)
	for i, b := range signalIDBytes {
		signalIDBytes[i] = ^b
	}

	signalIDBytesLen := len(signalIDBytes)

	key := make([]byte, 1+powerBytesLen+1+signalIDBytesLen)
	key[0] = FeedsByPowerIndexKeyPrefix[0]
	copy(key[1:powerBytesLen+1], powerBytes)
	key[powerBytesLen+1] = byte(signalIDBytesLen)
	copy(key[powerBytesLen+2:], signalIDBytes)

	return key
}
