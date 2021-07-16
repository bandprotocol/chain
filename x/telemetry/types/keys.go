package types

const (
	// ModuleName is the name of the module.
	ModuleName = "telemetry"
	// StoreKey to be used when creating the KVStore.
	StoreKey     = ModuleName
	RouterKey    = ModuleName
	QuerierRoute = ModuleName

	QueryTopBalances = "top_balances"
	DenomTag         = "denomTag "
)
