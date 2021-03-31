package types

const (
	// ModuleName is the name of the module.
	ModuleName = "coinswap"
	// StoreKey to be used when creating the KVStore.
	StoreKey          = ModuleName
	DefaultParamspace = ModuleName
	QuerierRoute      = ModuleName
)

var (
	InitialRateStoreKey = []byte("InitialRateStore") // key initial rate store
)
