package types

const (
	// module name
	ModuleName = "bandtss"

	// StoreKey to be used when creating the KVStore.
	StoreKey = ModuleName

	// RouterKey is the message route for the tss module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the tss module
	QuerierRoute = ModuleName
)

var (

	// GlobalStoreKeyPrefix is the prefix for global primitive state variables.
	GlobalStoreKeyPrefix = []byte{0x00}

	// ParamsKeyPrefix is a prefix for keys that store TSSMember's parameters
	ParamsKeyPrefix = []byte{0x01}
)
