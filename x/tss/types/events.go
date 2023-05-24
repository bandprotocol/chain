package types

const (
	EventTypeCreateGroup     = "create_group"
	EventTypeSubmitDKGRound1 = "submit_dkg_round1"
	EventTypeRound1Success   = "round1_success"

	AttributeKeyGroupID    = "group_id"
	AttributeKeyMemberID   = "member_id"
	AttributeKeyMember     = "member"
	AttributeKeySize       = "size"
	AttributeKeyThreshold  = "threshold"
	AttributeKeyPubKey     = "pub_key"
	AttributeKeyStatus     = "status"
	AttributeKeyDKGContext = "dkg_context"
	AttributeKeyRound1Data = "round1_data"
)
