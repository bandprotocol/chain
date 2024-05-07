package types

// events
const (
	EventTypeSubmitPrice          = "submit_price"
	EventTypeUpdatePrice          = "update_price"
	EventTypeSubmitSignal         = "submit_signal"
	EventTypeRemoveSignal         = "remove_signal"
	EventTypeCalculatePriceFailed = "calculate_price_failed"
	EventTypeUpdatePriceService   = "update_price_service"
	EventTypeUpdateParams         = "update_params"

	AttributeKeyPriceOption  = "price_option"
	AttributeKeyValidator    = "validator"
	AttributeKeyPrice        = "price"
	AttributeKeyTimestamp    = "timestamp"
	AttributeKeySignalID     = "signal_id"
	AttributeKeyDelegator    = "delegator"
	AttributeKeyPower        = "power"
	AttributeKeyErrorMessage = "error_message"
	AttributeKeyHash         = "hash"
	AttributeKeyVersion      = "version"
	AttributeKeyURL          = "url"
	AttributeKeyParams       = "params"
)
