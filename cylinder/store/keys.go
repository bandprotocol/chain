package store

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

var (
	// GlobalStoreKeyPrefix is the prefix for global primitive state variables.
	GlobalStoreKeyPrefix = []byte{0x00}
	// DECountStoreKey is the key that keeps the total DE count.
	DECountStoreKey = append(GlobalStoreKeyPrefix, []byte("DECount")...)

	// DKGStoreKeyPrefix is the prefix for DKG store.
	DKGStoreKeyPrefix = []byte{0x01}
	// GroupStoreKeyPrefix is the prefix for group store.
	GroupStoreKeyPrefix = []byte{0x02}
	// DEStoreKeyPrefix is the prefix for DE store.
	DEStoreKeyPrefix = []byte{0x03}
)

// DKGStoreKey returns the key to retrieve all data for a group.
func DKGStoreKey(groupID tss.GroupID) []byte {
	return append(DKGStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

// GroupStoreKey returns the key to retrieve all data for a group.
func GroupStoreKey(pubKey tss.Point) []byte {
	return append(GroupStoreKeyPrefix, pubKey...)
}

// DEStoreKey returns the key to retrieve private (d, e) by public (D, E).
func DEStoreKey(pubDE types.DE) []byte {
	bz := append(DEStoreKeyPrefix, pubDE.PubD...)
	return append(bz, pubDE.PubE...)
}
