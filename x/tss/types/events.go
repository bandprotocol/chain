package types

const (
	EventTypeCreateGroup     = "create_group"
	EventTypeSubmitDKGRound1 = "submit_dkg_round1"
	EventTypeRound1Success   = "round1_success"
	EventTypeSubmitDKGRound2 = "submit_dkg_round2"
	EventTypeRound2Success   = "round2_success"
	EventTypeComplainSuccess = "complain_success"
	EventTypeComplainFailed  = "complain_failed"
	EventTypeConfirmSuccess  = "confirm_success"
	EventTypeRound3Success   = "round3_success"
	EventTypeRound3Failed    = "round3_failed"
	EventTypeRequestSign     = "request_sign"
	EventTypeSignSuccess     = "sign_success"
	EventTypeSubmitSign      = "submit_sign"

	AttributeKeyGroupID              = "group_id"
	AttributeKeyMemberID             = "member_id"
	AttributeKeyMember               = "member"
	AttributeKeySize                 = "size"
	AttributeKeyThreshold            = "threshold"
	AttributeKeyPubKey               = "pub_key"
	AttributeKeyStatus               = "status"
	AttributeKeyDKGContext           = "dkg_context"
	AttributeKeyRound1Data           = "round1_data"
	AttributeKeyRound2Data           = "round2_data"
	AttributeKeyMemberIDI            = "member_id_i"
	AttributeKeyMemberIDJ            = "member_id_j"
	AttributeKeyKeySym               = "key_sym"
	AttributeKeySig                  = "sig"
	AttributeKeyComplain             = "complain"
	AttributeKeyGroupPubKey          = "group_pub_key"
	AttributeKeyOwnPubKeySig         = "own_pub_key_sig"
	AttributeKeySigningID            = "signing_id"
	AttributeKeyCommitment           = "commitment"
	AttributeKeyMessage              = "message"
	AttributeKeyGroupPubNonce        = "group_pub_nonce"
	AttributeKeyOwnPubNonces         = "own_pub_nonces"
	AttributeKeyPubD                 = "pub_d"
	AttributeKeyPubE                 = "pub_e"
	AttributeKeyAssignedParticipants = "assigned_participants"
)
