package types

const (
	// module name
	ModuleName = "rollingseed"

	// StoreKey to be used when creating the KVStore.
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the rollingseed module
	QuerierRoute = ModuleName
)

var (
	// RollingSeedSizeInBytes is the size of rolling block hash for random seed.
	RollingSeedSizeInBytes = 32

	// GlobalStoreKeyPrefix is the prefix for global primitive state variables.
	GlobalStoreKeyPrefix = []byte{0x00}

	// RollingSeedStoreKey is the key that keeps the seed based on the first 8-bit of the most recent 32 block hashes.
	RollingSeedStoreKey = append(GlobalStoreKeyPrefix, []byte("RollingSeed")...)
)
