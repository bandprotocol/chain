package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
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

// Constants for keys
var (
	// global store keys
	ReferenceSourceConfigStoreKey = []byte{0x00}
	CurrentFeedsStoreKey          = []byte{0x01}

	// store prefixes
	ValidatorPriceListStoreKeyPrefix = []byte{0x10}
	PriceStoreKeyPrefix              = []byte{0x11}
	VoteStoreKeyPrefix               = []byte{0x12}
	SignalTotalPowerStoreKeyPrefix   = []byte{0x13}

	// index prefixes
	SignalTotalPowerByPowerIndexKeyPrefix = []byte{0x80}

	// params store keys
	ParamsKey = []byte{0x90}
)

// VoteStoreKey creates a key for storing vote
func VoteStoreKey(voter sdk.AccAddress) []byte {
	return append(VoteStoreKeyPrefix, address.MustLengthPrefix(voter.Bytes())...)
}

// SignalTotalPowerStoreKey creates a key for storing signal-total-powers
func SignalTotalPowerStoreKey(signalID string) []byte {
	return append(SignalTotalPowerStoreKeyPrefix, []byte(signalID)...)
}

// ValidatorPriceListStoreKey creates a key for storing a validator prices list
func ValidatorPriceListStoreKey(validator sdk.ValAddress) []byte {
	return append(ValidatorPriceListStoreKeyPrefix, address.MustLengthPrefix(validator.Bytes())...)
}

// PriceStoreKey creates a key for storing price data
func PriceStoreKey(signalID string) []byte {
	return append(PriceStoreKeyPrefix, []byte(signalID)...)
}

// SignalTotalPowerByPowerIndexKey creates a key for storing signal-total-powers by power index
func SignalTotalPowerByPowerIndexKey(signalID string, power int64) []byte {
	powerBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(powerBytes, uint64(power))
	powerBytesLen := len(powerBytes) // 8

	signalIDBytes := []byte(signalID)
	for i, b := range signalIDBytes {
		signalIDBytes[i] = ^b
	}

	signalIDBytesLen := len(signalIDBytes)

	key := make([]byte, 1+powerBytesLen+1+signalIDBytesLen)
	key[0] = SignalTotalPowerByPowerIndexKeyPrefix[0]
	copy(key[1:powerBytesLen+1], powerBytes)
	key[powerBytesLen+1] = byte(signalIDBytesLen)
	copy(key[powerBytesLen+2:], signalIDBytes)

	return key
}
