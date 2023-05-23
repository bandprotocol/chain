package types

const (
	EventTypeCreateGroup     = "create_group"
	EventTypeSubmitDKGRound1 = "submit_dkg_round1"
	EventTypeRound1Success   = "round1_success"
	EventTypeSubmitDKGRound2 = "submit_dkg_round2"
	EventTypeRound2Success   = "round2_success"

	AttributeKeyGroupID            = "group_id"
	AttributeKeyMemberID           = "member_id"
	AttributeKeyMember             = "member"
	AttributeKeySize               = "size"
	AttributeKeyThreshold          = "threshold"
	AttributeKeyPubKey             = "pub_key"
	AttributeKeyStatus             = "status"
	AttributeKeyDKGContext         = "dkg_context"
	AttributeKeyCoefficientsCommit = "coefficients_commit"
	AttributeKeyOneTimePubKey      = "one_time_pub_key"
	AttributeKeyA0Sig              = "a0_sig"
	AttributeKeyOneTimeSig         = "one_time_sig"
	AttributeKeyRound2Data         = "round2_data"
)
