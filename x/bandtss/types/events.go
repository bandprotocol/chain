package types

const (
	EventTypeActivate         = "activate"
	EventTypeCreateGroup      = "create_group"
	EventTypeHealthCheck      = "health_check"
	EventTypeReplacement      = "replacement"
	EventTypeRequestSignature = "request_signature"
	EventTypeUpdateGroupFee   = "update_group_fee"
	EventTypeInactiveStatus   = "inactive_status"
	EventTypePausedStatus     = "paused_status"
	EventTypeJailStatus       = "jail_status"

	AttributeKeyAddress       = "address"
	AttributeKeyDKGContext    = "dkg_context"
	AttributeKeyFee           = "fee"
	AttributeKeyGroupID       = "group_id"
	AttributeKeyGroupPubNonce = "group_pub_nonce"
	AttributeKeyMemberID      = "member_id"
	AttributeKeyPubKey        = "pub_key"
	AttributeKeyReplacementID = "replacement_id"
	AttributeKeySigningID     = "signing_id"
	AttributeKeySize          = "size"
	AttributeKeyStatus        = "status"
	AttributeKeyThreshold     = "threshold"
)
