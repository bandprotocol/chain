package types

const (
	// ModuleName
	ModuleName = "mint"

	WrappedModuleName = "odin" + ModuleName

	// DefaultParamspace params keeper
	DefaultParamspace = ModuleName

	// StoreKey is the default store key for mint
	StoreKey = ModuleName

	// RouterKey is the message route for mint
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the minting store.
	QuerierRoute = StoreKey

	// Query endpoints supported by the minting querier
	QueryRoute                 = "minting"
	QueryParameters            = "parameters"
	QueryInflation             = "inflation"
	QueryAnnualProvisions      = "annual_provisions"
	QueryEthIntegrationAddress = "eth_integration_address"
	QueryTreasuryPool          = "treasury_pool"
)

var (
	// GlobalStoreKeyPrefix is used as prefix for the store keys
	GlobalStoreKeyPrefix = []byte{0x00}
	// MinterKey is used for the keeper store
	MinterKey = append(GlobalStoreKeyPrefix, []byte("Minter")...)
	// MintPoolStoreKey is the key for global mint pool state
	MintPoolStoreKey = append(GlobalStoreKeyPrefix, []byte("MintPool")...)
)
