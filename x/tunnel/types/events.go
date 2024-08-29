package types

// events
const (
	EventTypeUpdateParams            = "update_params"
	EventTypeCreateTunnel            = "create_tunnel"
	EventTypeActivateTunnel          = "activate_tunnel"
	EventTypeManualTriggerTunnel     = "manual_trigger_tunnel"
	EventTypeSignalInfoNotFound      = "signal_info_not_found"
	EventTypeProducePacketFail       = "produce_packet_fail"
	EventTypeGetSignalPricesInfoFail = "get_signal_prices_info_fail"

	AttributeKeyParams           = "params"
	AttributeKeyTunnelID         = "tunnel_id"
	AttributeKeySignalID         = "signal_id"
	AttributeKeyRoute            = "route"
	AttributeKeyFeedType         = "feed_type"
	AttributeKeyFeePayer         = "fee_payer"
	AttributeKeySignalPriceInfos = "signal_price_infos"
	AttributeKeyIsActive         = "is_active"
	AttributeKeyCreatedAt        = "created_at"
	AttributeKeyCreator          = "creator"
	AttributeKeyReason           = "reason"
)
