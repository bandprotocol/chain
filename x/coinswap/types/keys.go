package types

const (
	// ModuleName is the name of the module.
	ModuleName = "coinswap"
	// StoreKey to be used when creating the KVStore.
	StoreKey          = ModuleName
	DefaultParamspace = ModuleName
	QuerierRoute      = ModuleName

	QueryParams = "params"
	QueryRate   = "rate"
)

var (
	InitialRateStoreKey = []byte("InitialRateStore") // key initial rate store
)
