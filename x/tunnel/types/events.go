package types

// events
const (
	EventTypeUpdateParams        = "update_params"
	EventTypeCreateTunnel        = "create_tunnel"
	EventTypeEditTunnel          = "edit_tunnel"
	EventTypeActivateTunnel      = "activate_tunnel"
	EventTypeDeactivateTunnel    = "deactivate_tunnel"
	EventTypeManualTriggerTunnel = "manual_trigger_tunnel"
	EventTypeProducePacketFail   = "produce_packet_fail"

	AttributeKeyParams           = "params"
	AttributeKeyTunnelID         = "tunnel_id"
	AttributeKeyInterval         = "interval"
	AttributeKeyRoute            = "route"
	AttributeKeyEncoder          = "encoder"
	AttributeKeyInitialDeposit   = "initial_deposit"
	AttributeKeyFeePayer         = "fee_payer"
	AttributeKeySignalPriceInfos = "signal_price_infos"
	AttributeKeyIsActive         = "is_active"
	AttributeKeyCreatedAt        = "created_at"
	AttributeKeyCreator          = "creator"
	AttributeKeyReason           = "reason"
)
