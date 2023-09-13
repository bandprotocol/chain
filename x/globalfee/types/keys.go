package types

const (
	// ModuleName is the name of the this module
	ModuleName = "globalfee"

	QuerierRoute = ModuleName

	// StoreKey to be used when creating the KVStore.
	StoreKey = ModuleName
)

var (
	ParamsKeyPrefix = []byte{0x01}
)
