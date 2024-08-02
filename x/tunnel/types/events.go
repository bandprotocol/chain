package types

// events
const (
	EventTypeUpdateParams   = "update_params"
	EventTypeCreateTunnel   = "create_tunnel"
	EventTypeActivateTunnel = "activate_tunnel"
	EventTypeSendPacketFail = "send_packet_fail"

	AttributeKeyParams           = "params"
	AttributeKeyTunnelID         = "tunnel_id"
	AttributeKeyRoute            = "route"
	AttributeKeyFeedType         = "feed_type"
	AttributeKeyFeePayer         = "fee_payer"
	AttributeKeySignalPriceInfos = "signal_price_infos"
	AttributeKeyIsActive         = "is_active"
	AttributeKeyCreatedAt        = "created_at"
	AttributeKeyCreator          = "creator"
	AttributeKeyReason           = "reason"
)
