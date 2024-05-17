package types

// events
const (
	EventTypeSubmitPrice            = "submit_price"
	EventTypeUpdatePrice            = "update_price"
	EventTypeUpdateSignalTotalPower = "update_signal_total_power"
	EventTypeUpdateSupportedFeeds   = "update_supported_feeds"
	EventTypeCalculatePriceFailed   = "calculate_price_failed"
	EventTypeUpdatePriceService     = "update_price_service"
	EventTypeUpdateParams           = "update_params"

	AttributeKeyPriceStatus           = "price_status"
	AttributeKeyValidator             = "validator"
	AttributeKeyPrice                 = "price"
	AttributeKeyTimestamp             = "timestamp"
	AttributeKeySignalID              = "signal_id"
	AttributeKeyPower                 = "power"
	AttributeKeyInterval              = "interval"
	AttributeKeyLastUpdateTimestamp   = "last_update_timestamp"
	AttributeKeyLastUpdateBlock       = "last_update_block"
	AttributeKeyDeviationInThousandth = "deviation_in_thousandth"
	AttributeKeyErrorMessage          = "error_message"
	AttributeKeyHash                  = "hash"
	AttributeKeyVersion               = "version"
	AttributeKeyURL                   = "url"
	AttributeKeyParams                = "params"
)
