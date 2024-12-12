package types

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v3/pkg/tss"
)

var (
	_ sdk.Msg = &MsgSubmitDKGRound1{}
	_ sdk.Msg = &MsgSubmitDKGRound2{}
	_ sdk.Msg = &MsgComplain{}
	_ sdk.Msg = &MsgConfirm{}
	_ sdk.Msg = &MsgSubmitDEs{}
	_ sdk.Msg = &MsgResetDE{}
	_ sdk.Msg = &MsgSubmitSignature{}
	_ sdk.Msg = &MsgUpdateParams{}

	_ sdk.HasValidateBasic = (*MsgSubmitDKGRound1)(nil)
	_ sdk.HasValidateBasic = (*MsgSubmitDKGRound2)(nil)
	_ sdk.HasValidateBasic = (*MsgComplain)(nil)
	_ sdk.HasValidateBasic = (*MsgConfirm)(nil)
	_ sdk.HasValidateBasic = (*MsgSubmitDEs)(nil)
	_ sdk.HasValidateBasic = (*MsgResetDE)(nil)
	_ sdk.HasValidateBasic = (*MsgSubmitSignature)(nil)
	_ sdk.HasValidateBasic = (*MsgUpdateParams)(nil)
)

// ====================================
// MsgSubmitDKGRound1
// ====================================

// NewMsgSubmitDKGRound1 creates a new MsgSubmitDKGRound1 instance.
func NewMsgSubmitDKGRound1(groupID tss.GroupID, round1Info Round1Info, sender string) *MsgSubmitDKGRound1 {
	return &MsgSubmitDKGRound1{
		GroupID:    groupID,
		Round1Info: round1Info,
		Sender:     sender,
	}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgSubmitDKGRound1) ValidateBasic() error {
	// Validate group ID
	if m.GroupID == 0 {
		return ErrInvalidGroup.Wrap("group id cannot be 0")
	}

	// Validate member address
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	if err := m.Round1Info.Validate(); err != nil {
		return err
	}

	return nil
}

// ====================================
// MsgSubmitDKGRound2
// ====================================

// NewMsgSubmitDKGRound2 creates a new MsgSubmitDKGRound2 instance.
func NewMsgSubmitDKGRound2(groupID tss.GroupID, round2Info Round2Info, sender string) *MsgSubmitDKGRound2 {
	return &MsgSubmitDKGRound2{
		GroupID:    groupID,
		Round2Info: round2Info,
		Sender:     sender,
	}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgSubmitDKGRound2) ValidateBasic() error {
	// Validate group ID
	if m.GroupID == 0 {
		return ErrInvalidGroup.Wrap("group id cannot be 0")
	}

	// Validate member address
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	if err := m.Round2Info.Validate(); err != nil {
		return err
	}

	return nil
}

// ====================================
// MsgComplain
// ====================================

// NewMsgComplain creates a new MsgComplain instance.
func NewMsgComplain(groupID tss.GroupID, complaints []Complaint, sender string) *MsgComplain {
	return &MsgComplain{
		GroupID:    groupID,
		Complaints: complaints,
		Sender:     sender,
	}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgComplain) ValidateBasic() error {
	// Validate group ID
	if m.GroupID == 0 {
		return ErrInvalidGroup.Wrap("group id cannot be 0")
	}

	// Validate member address
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	// Validate complaints size
	if len(m.Complaints) == 0 {
		return ErrInvalidComplaint.Wrapf("must contain at least one complaint")
	}

	// Validate complaints
	memberI := m.Complaints[0].Complainant
	for i, c := range m.Complaints {
		// Validate member complainant
		if i > 0 && memberI != c.Complainant {
			return ErrInvalidComplaint.Wrapf("memberID complainant in the list of complaints must be the same value")
		}

		// Validate complaints
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ====================================
// MsgConfirm
// ====================================

// NewMsgConfirm creates a new MsgConfirm instance.
func NewMsgConfirm(
	groupID tss.GroupID,
	memberID tss.MemberID,
	ownPubKeySig tss.Signature,
	sender string,
) *MsgConfirm {
	return &MsgConfirm{
		GroupID:      groupID,
		MemberID:     memberID,
		OwnPubKeySig: ownPubKeySig,
		Sender:       sender,
	}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgConfirm) ValidateBasic() error {
	// Validate member ID
	if m.MemberID == 0 {
		return ErrInvalidMember.Wrap("member id cannot be 0")
	}

	// Validate group ID
	if m.GroupID == 0 {
		return ErrInvalidGroup.Wrap("group id cannot be 0")
	}

	// Validate own pub key sig
	if err := m.OwnPubKeySig.Validate(); err != nil {
		return ErrInvalidPublicKey.Wrapf("invalid own public key signature: %s", err)
	}

	// Validate member address
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	return nil
}

// ====================================
// MsgSubmitDEs
// ====================================

// NewMsgSubmitDEs creates a new MsgSubmitDEs instance.
func NewMsgSubmitDEs(des []DE, sender string) *MsgSubmitDEs {
	return &MsgSubmitDEs{
		DEs:    des,
		Sender: sender,
	}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgSubmitDEs) ValidateBasic() error {
	// Validate member address
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	// Validate DEs
	for i, de := range m.DEs {
		if err := de.Validate(); err != nil {
			return errorsmod.Wrapf(err, "DE index %d", i)
		}
	}

	return nil
}

// ====================================
// MsgResetDE
// ====================================

// NewMsgResetDE creates a new MsgResetDE instance.
func NewMsgResetDE(sender string) *MsgResetDE {
	return &MsgResetDE{
		Sender: sender,
	}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgResetDE) ValidateBasic() error {
	// Validate member address
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	return nil
}

// ====================================
// MsgSubmitSignature
// ====================================

// NewMsgSubmitSignature creates a new MsgSubmitSignature instance.
func NewMsgSubmitSignature(
	signingID tss.SigningID,
	memberID tss.MemberID,
	signature tss.Signature,
	signer string,
) *MsgSubmitSignature {
	return &MsgSubmitSignature{
		SigningID: signingID,
		MemberID:  memberID,
		Signature: signature,
		Signer:    signer,
	}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgSubmitSignature) ValidateBasic() error {
	// Validate member ID
	if m.SigningID == 0 {
		return ErrInvalidSigning.Wrap("signing id cannot be 0")
	}

	// Validate member ID
	if m.MemberID == 0 {
		return ErrInvalidMember.Wrap("member id cannot be 0")
	}

	// Validate member address
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid signer address: %s", err)
	}

	// Validate member signature
	if err := m.Signature.Validate(); err != nil {
		return ErrInvalidSignature.Wrapf("invalid signature :%s", err)
	}

	return nil
}

// ====================================
// NewMsgUpdateParams
// ====================================

// NewMsgUpdateParams creates a new MsgUpdateParams instance
func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	if err := m.Params.Validate(); err != nil {
		return err
	}

	return nil
}
