package types

const (
	// ModuleName defines the module name
	ModuleName = "restake"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the restake module
	QuerierRoute = ModuleName
)

var (
	GlobalStoreKeyPrefix = []byte{0x00}
)
