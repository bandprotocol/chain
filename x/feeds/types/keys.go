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

	SymbolStoreKeyPrefix          = []byte{0x01}
	PriceValidatorStoreKeyPrefix  = []byte{0x02}
	PriceStoreKeyPrefix           = []byte{0x03}
	DelegatorSignalStoreKeyPrefix = []byte{0x04}

	ParamsKey = []byte{0x10}

	SymbolsByPowerIndexKey = []byte{0x20}
)

func DelegatorSignalStoreKey(delegator sdk.AccAddress) []byte {
	return append(DelegatorSignalStoreKeyPrefix, delegator...)
}

func SymbolStoreKey(symbol string) []byte {
	return append(SymbolStoreKeyPrefix, []byte(symbol)...)
}

func PriceValidatorsStoreKey(symbol string) []byte {
	return append(PriceValidatorStoreKeyPrefix, []byte(symbol)...)
}

func PriceValidatorStoreKey(symbol string, validator sdk.ValAddress) []byte {
	return append(PriceValidatorsStoreKey(symbol), validator...)
}

func PriceStoreKey(symbol string) []byte {
	return append(PriceStoreKeyPrefix, []byte(symbol)...)
}

func GetSymbolsByPowerIndexKey(symbol string, power uint64) []byte {
	powerBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(powerBytes, power)
	powerBytesLen := len(powerBytes) // 8

	symbolBytes := []byte(symbol)
	for i, b := range symbolBytes {
		symbolBytes[i] = ^b
	}

	symbolBytesLen := len(symbolBytes)

	key := make([]byte, 1+powerBytesLen+1+symbolBytesLen)
	key[0] = SymbolsByPowerIndexKey[0]
	copy(key[1:powerBytesLen+1], powerBytes)
	key[powerBytesLen+1] = byte(symbolBytesLen)
	copy(key[powerBytesLen+2:], symbolBytes)

	return key
}
