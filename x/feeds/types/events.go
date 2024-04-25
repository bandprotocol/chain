package types

// events
const (
	EventTypeSubmitPrice          = "submit_price"
	EventTypeUpdatePrice          = "update_price"
	EventTypeSubmitSignals        = "submit_signals"
	EventTypeRemoveSignals        = "remove_signals"
	EventTypeCalculatePriceFailed = "calculate_price_failed"

	AttributeKeyPriceOption  = "price_option"
	AttributeKeyValidator    = "validator"
	AttributeKeyPrice        = "price"
	AttributeKeyTimestamp    = "timestamp"
	AttributeKeySignalID     = "signal_id"
	AttributeKeyDelegator    = "delegator"
	AttributeKeyPower        = "power"
	AttributeKeyErrorMessage = "error_message"
)
