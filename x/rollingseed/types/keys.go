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

	// global store keys
	RollingSeedStoreKey = []byte{0x00}
)
