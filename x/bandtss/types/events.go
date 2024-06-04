package types

const (
	EventTypeActivate              = "activate"
	EventTypeHealthCheck           = "health_check"
	EventTypeInactiveStatus        = "inactive_status"
	EventTypeReplacement           = "replacement"
	EventTypeSigningRequestCreated = "bandtss_signing_request_created"
	EventTypeNewGroupActivate      = "new_group_activate"

	AttributeKeyAddress                 = "address"
	AttributeKeySigningID               = "bandtss_signing_id"
	AttributeKeyCurrentGroupID          = "current_group_id"
	AttributeKeyReplacingGroupID        = "replacing_group_id"
	AttributeKeyCurrentGroupSigningID   = "current_group_signing_id"
	AttributeKeyReplacingGroupSigningID = "replacing_group_signing_id"
	AttributeKeyReplacementStatus       = "replacement_status"
	AttributeKeyExecTime                = "exec_time"
	AttributeKeyGroupID                 = "group_id"
	AttributeKeyGroupPubKey             = "group_pub_key"
	AttributeKeyNewGroupPubKey          = "new_group_pub_key"
	AttributeKeyRAddress                = "r_address"
	AttributeKeySignature               = "signature"
)
