package types

const (
	EventTypeCreateGroup        = "create_group"
	EventTypeUpdateGroupFee     = "update_group_fee"
	EventTypeReplacement        = "replacement"
	EventTypeExpiredGroup       = "expired_group"
	EventTypeSubmitDKGRound1    = "submit_dkg_round1"
	EventTypeRound1Success      = "round1_success"
	EventTypeSubmitDKGRound2    = "submit_dkg_round2"
	EventTypeRound2Success      = "round2_success"
	EventTypeComplainSuccess    = "complain_success"
	EventTypeComplainFailed     = "complain_failed"
	EventTypeConfirmSuccess     = "confirm_success"
	EventTypeRound3Success      = "round3_success"
	EventTypeRound3Failed       = "round3_failed"
	EventTypeRequestSignature   = "request_signature"
	EventTypeSigningSuccess     = "signing_success"
	EventTypeSigningFailed      = "signing_failed"
	EventTypeExpiredSigning     = "expired_signing"
	EventTypeReplacementSuccess = "replacement_success"
	EventTypeReplacementFailed  = "replacement_failed"
	EventTypeSubmitSignature    = "submit_signature"
	EventTypeActivate           = "activate"
	EventTypeHealthCheck        = "health_check"
	EventTypeInactive           = "inactive"

	AttributeKeyGroupID       = "group_id"
	AttributeKeyReplacementID = "replacement_id"
	AttributeKeyMemberID      = "member_id"
	AttributeKeyAddress       = "address"
	AttributeKeySize          = "size"
	AttributeKeyThreshold     = "threshold"
	AttributeKeyPubKey        = "pub_key"
	AttributeKeyStatus        = "status"
	AttributeKeyFee           = "fee"
	AttributeKeyDKGContext    = "dkg_context"
	AttributeKeyRound1Info    = "round1_info"
	AttributeKeyRound2Info    = "round2_info"
	AttributeKeyComplainantID = "complainant_id"
	AttributeKeyRespondentID  = "respondent_id"
	AttributeKeyKeySym        = "key_sym"
	AttributeKeySignature     = "signature"
	AttributeKeyGroupPubKey   = "group_pub_key"
	AttributeKeyOwnPubKeySig  = "own_pub_key_sig"
	AttributeKeySigningID     = "signing_id"
	AttributeKeyReason        = "reason"
	AttributeKeyMessage       = "message"
	AttributeKeyGroupPubNonce = "group_pub_nonce"
	AttributeKeyPubNonce      = "pub_nonce"
	AttributeKeyBindingFactor = "binding_factor"
	AttributeKeyPubD          = "pub_d"
	AttributeKeyPubE          = "pub_e"
	AttributeKeyFromGroupID   = "from_group_id"
	AttributeKeyToGroupID     = "to_group_id"
)