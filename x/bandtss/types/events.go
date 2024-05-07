package types

const (
	EventTypeActivate              = "activate"
	EventTypeHealthCheck           = "health_check"
	EventTypeInactiveStatus        = "inactive_status"
	EventTypeFirstGroupCreated     = "first_group_created"
	EventTypeReplacement           = "replacement"
	EventTypeSigningRequestCreated = "bandtss_signing_request_created"

	AttributeKeyAddress                 = "address"
	AttributeKeySigningID               = "bandtss_signing_id"
	AttributeKeyCurrentGroupID          = "current_group_id"
	AttributeKeyReplacingGroupID        = "replacing_group_id"
	AttributeKeyCurrentGroupSigningID   = "current_group_signing_id"
	AttributeKeyReplacingGroupSigningID = "replacing_group_signing_id"
	AttributeKeyReplacementStatus       = "replacement_status"
)
