package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName defines the module name
	ModuleName = "feed"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the feed module
	QuerierRoute = ModuleName
)

var (
	GlobalStoreKeyPrefix = []byte{0x00}

	SymbolStoreKeyPrefix         = []byte{0x01}
	PriceValidatorStoreKeyPrefix = []byte{0x02}
	PriceStoreKeyPrefix          = []byte{0x03}

	ParamsKey = []byte{0x04}
)

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
