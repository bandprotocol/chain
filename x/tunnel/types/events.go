package types

// events
const (
	EventTypeUpdateParams         = "update_params"
	EventTypeCreateTunnel         = "create_tunnel"
	EventTypeUpdateAndResetTunnel = "update_and_reset_tunnel"
	EventTypeActivateTunnel       = "activate_tunnel"
	EventTypeDeactivateTunnel     = "deactivate_tunnel"
	EventTypeTriggerTunnel        = "trigger_tunnel"
	EventTypeProducePacketFail    = "produce_packet_fail"
	EventTypeProducePacketSuccess = "produce_packet_success"

	AttributeKeyParams          = "params"
	AttributeKeyTunnelID        = "tunnel_id"
	AttributeKeySequence        = "sequence"
	AttributeKeyInterval        = "interval"
	AttributeKeyRoute           = "route"
	AttributeKeyEncoder         = "encoder"
	AttributeKeyInitialDeposit  = "initial_deposit"
	AttributeKeyFeePayer        = "fee_payer"
	AttributeKeySignalDeviation = "signal_deviation"
	AttributeKeyIsActive        = "is_active"
	AttributeKeyCreatedAt       = "created_at"
	AttributeKeyCreator         = "creator"
	AttributeKeyReason          = "reason"
)
