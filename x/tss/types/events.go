package types

const (
	EventTypeCreateGroup      = "create_group"
	EventTypeSubmitDKGRound1  = "submit_dkg_round1"
	EventTypeRound1Success    = "round1_success"
	EventTypeSubmitDKGRound2  = "submit_dkg_round2"
	EventTypeRound2Success    = "round2_success"
	EventTypeComplainsSuccess = "complains_success"
	EventTypeComplainsFailed  = "complains_failed"
	EventTypeRound3Success    = "round3_success"

	AttributeKeyGroupID    = "group_id"
	AttributeKeyMemberID   = "member_id"
	AttributeKeyMember     = "member"
	AttributeKeySize       = "size"
	AttributeKeyThreshold  = "threshold"
	AttributeKeyPubKey     = "pub_key"
	AttributeKeyStatus     = "status"
	AttributeKeyDKGContext = "dkg_context"
	AttributeKeyRound1Data = "round1_data"
	AttributeKeyRound2Data = "round2_data"
	AttributeKeyComplains  = "complains"
	AttributeOwnPubKeySig  = "own_pub_key_sig"
)
