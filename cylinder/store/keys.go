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

// ReportStoreKey returns the key to retrieve all data reports for a request.
func GroupStoreKey(groupID tss.GroupID) []byte {
	return append(GroupStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
}

// ReportStoreKey returns the key to retrieve all data reports for a request.
func DEStoreKey(D uint64, E uint64) []byte {
	bz := append(DEStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(D))...)
	return append(bz, sdk.Uint64ToBigEndian(E)...)
}
