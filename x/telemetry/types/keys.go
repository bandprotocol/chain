package types

const (
	// ModuleName is the name of the module.
	ModuleName = "telemetry"
	// StoreKey to be used when creating the KVStore.
	StoreKey     = ModuleName
	RouterKey    = ModuleName
	QuerierRoute = ModuleName

	QueryTopBalances        = "top_balances"
	QueryExtendedValidators = "extended_validators"
	QueryAvgBlockSize       = "avg_block_size"
	QueryAvgBlockTime       = "avg_block_time"
	QueryAvgTxFee           = "avg_tx_fee"
	QueryTxVolume           = "tx_volume"
	QueryValidatorsBlocks   = "validators_blocks"

	DenomTag  = "denom"
	StatusTag = "status"
)
