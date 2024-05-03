package types

const (
	EventTypeActivate           = "activate"
	EventTypeHealthCheck        = "health_check"
	EventTypeInactiveStatus     = "inactive_status"
	EventTypeReplacement        = "replacement"
	EventTypeReplacementSuccess = "replacement_success"
	EventTypeReplacementFailed  = "replacement_failed"

	AttributeKeyAddress          = "address"
	AttributeKeyCurrentGroupID   = "current_group_id"
	AttributeKeyReplacingGroupID = "replacing_group_id"
)
