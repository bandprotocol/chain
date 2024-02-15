package types

const (
	// ModuleName is the name of the this module
	ModuleName = "globalfee"

	// QuerierRoute is the querier route for the globalfee module
	QuerierRoute = ModuleName

	// StoreKey to be used when creating the KVStore.
	StoreKey = ModuleName

	// RouterKey is the msg router key for the globalfee module
	RouterKey = ModuleName
)

var ParamsKeyPrefix = []byte{0x01}
