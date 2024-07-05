package types

// events
const (
	EventTypeSubmitSignalPrice           = "submit_signal_price"
	EventTypeUpdatePrice                 = "update_price"
	EventTypeUpdateSignalTotalPower      = "update_signal_total_power"
	EventTypeDeleteSignalTotalPower      = "delete_signal_total_power"
	EventTypeUpdateSupportedFeeds        = "update_supported_feeds"
	EventTypeCalculatePriceFailed        = "calculate_price_failed"
	EventTypeUpdateReferenceSourceConfig = "update_reference_source_config"
	EventTypeUpdateParams                = "update_params"

	AttributeKeyPriceStatus         = "price_status"
	AttributeKeyValidator           = "validator"
	AttributeKeyPrice               = "price"
	AttributeKeyTimestamp           = "timestamp"
	AttributeKeySignalID            = "signal_id"
	AttributeKeyPower               = "power"
	AttributeKeyInterval            = "interval"
	AttributeKeyLastUpdateTimestamp = "last_update_timestamp"
	AttributeKeyLastUpdateBlock     = "last_update_block"
	AttributeKeyDeviationBasisPoint = "deviation_basis_point"
	AttributeKeyErrorMessage        = "error_message"
	AttributeKeyIPFSHash            = "ipfs_hash"
	AttributeKeyVersion             = "version"
	AttributeKeyParams              = "params"
)
