package types

// events
const (
	EventTypeSubmitPrice          = "submit_price"
	EventTypeUpdatePrice          = "update_price"
	EventTypeUpdateFeed           = "update_feed"
	EventTypeDeleteFeed           = "delete_feed"
	EventTypeCalculatePriceFailed = "calculate_price_failed"
	EventTypeUpdatePriceService   = "update_price_service"
	EventTypeUpdateParams         = "update_params"

	AttributeKeyPriceStatus                 = "price_status"
	AttributeKeyValidator                   = "validator"
	AttributeKeyPrice                       = "price"
	AttributeKeyTimestamp                   = "timestamp"
	AttributeKeySignalID                    = "signal_id"
	AttributeKeyPower                       = "power"
	AttributeKeyInterval                    = "interval"
	AttributeKeyLastIntervalUpdateTimestamp = "last_interval_update_timestamp"
	AttributeKeyDeviationInThousandth       = "deviation_in_thousandth"
	AttributeKeyErrorMessage                = "error_message"
	AttributeKeyHash                        = "hash"
	AttributeKeyVersion                     = "version"
	AttributeKeyURL                         = "url"
	AttributeKeyParams                      = "params"
)
