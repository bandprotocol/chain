package store

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// GlobalStoreKeyPrefix is the prefix for global primitive state variables.
	GlobalStoreKeyPrefix = []byte{0x00}
	// DECountStoreKey is the key that keeps the total DE count.
	DECountStoreKey = append(GlobalStoreKeyPrefix, []byte("DECount")...)

	// GroupStoreKeyPrefix is the prefix for group store.
	GroupStoreKeyPrefix = []byte{0x01}
	// DEStoreKeyPrefix is the prefix for DE store.
	DEStoreKeyPrefix = []byte{0x02}
)

// GroupStoreKey returns the key to retrieve all data for a group.
func GroupStoreKey(groupID tss.GroupID) []byte {
	return append(GroupStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

// DEStoreKey returns the key to retrieve private (d, e) by public (D, E).
func DEStoreKey(pubD tss.PublicKey, pubE tss.PublicKey) []byte {
	bz := append(DEStoreKeyPrefix, pubD...)
	return append(bz, pubE...)
}
