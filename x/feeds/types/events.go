package types

// events
const (
	EventTypeUpdateSymbol = "update_symbol"
	EventTypeRemoveSymbol = "remove_symbol"
	EventTypeSubmitPrice  = "submit_price"
	EventTypeUpdatePrice  = "update_price"

	AttributeKeyValidator   = "validator"
	AttributeKeyPrice       = "price"
	AttributeKeyTimestamp   = "timestamp"
	AttributeKeySymbol      = "symbol"
	AttributeKeyMinInterval = "min_interval"
	AttributeKeyMaxInterval = "max_interval"
)
