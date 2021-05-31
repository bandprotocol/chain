package types

const (
	// ModuleName is the name of the module.
	ModuleName = "auction"
	// StoreKey to be used when creating the KVStore.
	StoreKey     = ModuleName
	QuerierRoute = ModuleName

	QueryParams        = "params"
	QueryAuctionStatus = "status"
)

var (
	// GlobalStoreKeyPrefix is the prefix for global primitive state variables.
	GlobalStoreKeyPrefix = []byte{0x00}
	// AuctionStatusStoreKey is the key that keeps auction status
	AuctionStatusStoreKey = append(GlobalStoreKeyPrefix, []byte("AuctionStatus")...)
)
