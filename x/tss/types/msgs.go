package types

import (
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
	_ sdk.Msg = &MsgSubmitSignature{}
	_ sdk.Msg = &MsgUpdateParams{}

	_ sdk.HasValidateBasic = (*MsgSubmitDKGRound1)(nil)
	_ sdk.HasValidateBasic = (*MsgSubmitDKGRound2)(nil)
	_ sdk.HasValidateBasic = (*MsgComplain)(nil)
	_ sdk.HasValidateBasic = (*MsgConfirm)(nil)
	_ sdk.HasValidateBasic = (*MsgSubmitDEs)(nil)
	_ sdk.HasValidateBasic = (*MsgSubmitSignature)(nil)
	_ sdk.HasValidateBasic = (*MsgUpdateParams)(nil)
)

// NewMsgSubmitDKGRound1 creates a new MsgSubmitDKGRound1 instance.
func NewMsgSubmitDKGRound1(groupID tss.GroupID, round1Info Round1Info, sender string) *MsgSubmitDKGRound1 {
	return &MsgSubmitDKGRound1{
		GroupID:    groupID,
		Round1Info: round1Info,
		Sender:     sender,
	}
}

// Type returns message type name.
func (m MsgSubmitDKGRound1) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners returns the expected signers for a MsgSubmitDKGRound1.
func (m MsgSubmitDKGRound1) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Sender)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgSubmitDKGRound1) ValidateBasic() error {
	// Validate member address
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	// Validate coefficients commit
	for _, c := range m.Round1Info.CoefficientCommits {
		if err := c.Validate(); err != nil {
			return ErrInvalidCoefficientCommit.Wrapf("invalid coefficient commit: %s", err)
		}
	}

	// Validate one time pub key
	if err := m.Round1Info.OneTimePubKey.Validate(); err != nil {
		return ErrInvalidPublicKey.Wrapf("invalid one-time public key: %s", err)
	}

	// Validate a0 signature
	if err := m.Round1Info.A0Signature.Validate(); err != nil {
		return ErrInvalidSignature.Wrapf("invalid a0 signature: %s", err)
	}

	// Validate one time signature
	if err := m.Round1Info.OneTimeSignature.Validate(); err != nil {
		return ErrInvalidSignature.Wrapf("invalid one-time signature: %s", err)
	}

	return nil
}

// NewMsgSubmitDKGRound2 creates a new MsgSubmitDKGRound2 instance.
func NewMsgSubmitDKGRound2(groupID tss.GroupID, round2Info Round2Info, sender string) *MsgSubmitDKGRound2 {
	return &MsgSubmitDKGRound2{
		GroupID:    groupID,
		Round2Info: round2Info,
		Sender:     sender,
	}
}

// Type returns message type name.
func (m MsgSubmitDKGRound2) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners returns the expected signers for a MsgSubmitDKGRound2.
func (m MsgSubmitDKGRound2) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Sender)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgSubmitDKGRound2) ValidateBasic() error {
	// Validate member address
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	// Validate encrypted secret shares
	for i, ess := range m.Round2Info.EncryptedSecretShares {
		if err := ess.Validate(); err != nil {
			return ErrInvalidSecretShare.Wrapf("encrypted secret shares at index %d: %s", i, err)
		}
	}

	return nil
}

// NewMsgComplain creates a new MsgComplain instance.
func NewMsgComplain(groupID tss.GroupID, complaints []Complaint, sender string) *MsgComplain {
	return &MsgComplain{
		GroupID:    groupID,
		Complaints: complaints,
		Sender:     sender,
	}
}

// Type returns message type name.
func (m MsgComplain) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners returns the expected signers for a MsgComplain.
func (m MsgComplain) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Sender)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgComplain) ValidateBasic() error {
	// Validate member address
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	// Validate complaints size
	if len(m.Complaints) < 1 {
		return ErrInvalidComplaint.Wrapf("must contain at least one complaint")
	}

	// Validate complaints
	memberI := m.Complaints[0].Complainant
	for i, c := range m.Complaints {
		// Validate member complainant
		if i > 0 && memberI != c.Complainant {
			return ErrInvalidComplaint.Wrapf("memberID complainant in the list of complaints must be the same value")
		}

		// Validate member complainant and respondent
		if c.Complainant == c.Respondent {
			return ErrInvalidComplaint.Wrapf("memberID complainant and respondent can not be the same value")
		}

		// Validate key sym
		if err := c.KeySym.Validate(); err != nil {
			return ErrInvalidSymmetricKey.Wrapf("invalid symmetric key: %s", err)
		}

		// Validate signature
		if err := c.Signature.Validate(); err != nil {
			return ErrInvalidSignature.Wrapf("invalid signature: %s", err)
		}
	}

	return nil
}

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

// Type returns message type name.
func (m MsgConfirm) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners returns the expected signers for a MsgConfirm.
func (m MsgConfirm) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Sender)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgConfirm) ValidateBasic() error {
	// Validate member address
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	// Validate own pub key sig
	if err := m.OwnPubKeySig.Validate(); err != nil {
		return ErrInvalidPublicKey.Wrapf("invalid own public key signature: %s", err)
	}

	return nil
}

// NewMsgSubmitDEs creates a new MsgSubmitDEs instance.
func NewMsgSubmitDEs(des []DE, sender string) *MsgSubmitDEs {
	return &MsgSubmitDEs{
		DEs:    des,
		Sender: sender,
	}
}

// Type returns message type name.
func (m MsgSubmitDEs) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners returns the expected signers for a MsgSubmitDEs.
func (m MsgSubmitDEs) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Sender)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgSubmitDEs) ValidateBasic() error {
	// Validate member address
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	// Validate DEs
	for i, de := range m.DEs {
		// Validate public key D
		if err := de.PubD.Validate(); err != nil {
			return ErrInvalidDE.Wrapf("pub D in DE index %d: %s", i, err)
		}

		// Validate public key E
		if err := de.PubE.Validate(); err != nil {
			return ErrInvalidDE.Wrapf("pub E in DE index %d: %s", i, err)
		}
	}

	return nil
}

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

// Type returns message type name.
func (m MsgSubmitSignature) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners returns the expected signers for a MsgSubmitSignature.
func (m MsgSubmitSignature) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Signer)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgSubmitSignature) ValidateBasic() error {
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

// NewMsgUpdateParams creates a new MsgUpdateParams instance
func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// Type returns message type name.
func (m MsgUpdateParams) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Authority)}
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
