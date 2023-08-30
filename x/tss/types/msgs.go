package types

import (
	"fmt"

	"cosmossdk.io/errors"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
)

var (
	_, _, _, _, _, _ sdk.Msg                       = &MsgCreateGroup{}, &MsgSubmitDKGRound1{}, &MsgSubmitDKGRound2{}, &MsgComplain{}, &MsgConfirm{}, &MsgSubmitDEs{}
	_, _, _, _, _, _ sdk.Msg                       = &MsgRequestSignature{}, &MsgSubmitSignature{}, &MsgActivate{}, &MsgActive{}, &MsgReplaceGroup{}, &MsgUpdateParams{}
	_                types.UnpackInterfacesMessage = &MsgRequestSignature{}
)

// Route Implements Msg.
func (m MsgCreateGroup) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgCreateGroup) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgCreateGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateGroup.
func (m MsgCreateGroup) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Authority)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgCreateGroup) ValidateBasic() error {
	// Validate members address
	for _, member := range m.Members {
		_, err := sdk.AccAddressFromBech32(member)
		if err != nil {
			return errors.Wrap(
				err,
				fmt.Sprintf("member: %s ", member),
			)
		}
	}

	// Check duplicate member
	if DuplicateInArray(m.Members) {
		return errors.Wrap(fmt.Errorf("members can not duplicate"), "members")
	}

	// Validate sender address
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf("sender: %s", m.Authority),
		)
	}

	// Validate threshold must be less than or equal to members but more than zero
	if m.Threshold > uint64(len(m.Members)) || m.Threshold <= 0 {
		return errors.Wrap(
			fmt.Errorf("threshold must be less than or equal to the members but more than zero"),
			"threshold",
		)
	}

	return nil
}

// Route Implements Msg.
func (m MsgReplaceGroup) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgReplaceGroup) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgReplaceGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateGroup.
func (m MsgReplaceGroup) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Authority)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgReplaceGroup) ValidateBasic() error {
	// Validate sender address
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf("sender: %s", m.Authority),
		)
	}

	return nil
}

// Route Implements Msg.
func (m MsgSubmitDKGRound1) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgSubmitDKGRound1) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgSubmitDKGRound1) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgSubmitDKGRound1.
func (m MsgSubmitDKGRound1) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Member)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgSubmitDKGRound1) ValidateBasic() error {
	// Validate member address
	_, err := sdk.AccAddressFromBech32(m.Member)
	if err != nil {
		return errors.Wrap(err, "member")
	}

	// Validate coefficients commit
	for _, c := range m.Round1Info.CoefficientCommits {
		if err := c.Validate(); err != nil {
			return errors.Wrap(err, "coefficients commit")
		}
	}

	// Validate one time pub key
	if err := m.Round1Info.OneTimePubKey.Validate(); err != nil {
		return errors.Wrap(err, "one time pub key")
	}

	// Validate a0 signature
	if err := m.Round1Info.A0Sig.Validate(); err != nil {
		return errors.Wrap(err, "a0 sig")
	}

	// Validate one time signature
	if err := m.Round1Info.OneTimeSig.Validate(); err != nil {
		return errors.Wrap(err, "one time sig")
	}

	return nil
}

// Route Implements Msg.
func (m MsgSubmitDKGRound2) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgSubmitDKGRound2) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgSubmitDKGRound2) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgSubmitDKGRound2.
func (m MsgSubmitDKGRound2) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Member)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgSubmitDKGRound2) ValidateBasic() error {
	// Validate member address
	_, err := sdk.AccAddressFromBech32(m.Member)
	if err != nil {
		return errors.Wrap(err, "member")
	}

	// Validate encrypted secret shares
	for _, ess := range m.Round2Info.EncryptedSecretShares {
		if err := ess.Validate(); err != nil {
			return errors.Wrap(err, "encrypted secret shares")
		}
	}

	return nil
}

// Route Implements Msg.
func (m MsgComplain) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgComplain) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgComplain) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgComplain.
func (m MsgComplain) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Member)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgComplain) ValidateBasic() error {
	// Validate member address
	_, err := sdk.AccAddressFromBech32(m.Member)
	if err != nil {
		return errors.Wrap(err, "member")
	}

	// Validate complaints size
	if len(m.Complaints) < 1 {
		return errors.Wrap(fmt.Errorf("must contain at least one complaint"), "complaints")
	}

	// Validate complaints
	memberI := m.Complaints[0].Complainant
	for i, c := range m.Complaints {
		// Validate member complainant
		if i > 0 && memberI != c.Complainant {
			return errors.Wrap(
				fmt.Errorf("memberID complainant in the list of complaints must be the same value"),
				"complainant",
			)
		}

		// Validate member complainant and respondent
		if c.Complainant == c.Respondent {
			return errors.Wrap(
				fmt.Errorf("memberID complainant and respondent can not be the same value"),
				"complainant, respondent",
			)
		}

		// Validate key sym
		if err := c.KeySym.Validate(); err != nil {
			return errors.Wrap(err, "key sym")
		}

		// Validate signature
		if err := c.Signature.Validate(); err != nil {
			return errors.Wrap(err, "signature")
		}
	}

	return nil
}

// Route Implements Msg.
func (m MsgConfirm) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgConfirm) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgConfirm) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgConfirm.
func (m MsgConfirm) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Member)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgConfirm) ValidateBasic() error {
	// Validate member address
	_, err := sdk.AccAddressFromBech32(m.Member)
	if err != nil {
		return errors.Wrap(err, "member")
	}

	// Validate own pub key sig
	if err = m.OwnPubKeySig.Validate(); err != nil {
		return errors.Wrap(err, "own pub key sig")
	}

	return nil
}

// Route Implements Msg.
func (m MsgSubmitDEs) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgSubmitDEs) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgSubmitDEs) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgSubmitDEs.
func (m MsgSubmitDEs) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Member)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgSubmitDEs) ValidateBasic() error {
	// Validate member address
	_, err := sdk.AccAddressFromBech32(m.Member)
	if err != nil {
		return errors.Wrap(err, "member")
	}

	// Validate DEs
	for i, de := range m.DEs {
		// Validate public key D
		if err = de.PubD.Validate(); err != nil {
			return errors.Wrap(err, fmt.Sprintf("pub D in DE index: %d", i))
		}

		// Validate public key E
		if err = de.PubE.Validate(); err != nil {
			return errors.Wrap(err, fmt.Sprintf("pub E in DE index: %d", i))
		}
	}

	return nil
}

// NewMsgRequestSignature creates a new MsgRequestSignature.
//
//nolint:interfacer
func NewMsgRequestSignature(
	gid tss.GroupID,
	content Content,
	feeLimit sdk.Coins,
	sender sdk.AccAddress,
) (*MsgRequestSignature, error) {
	m := &MsgRequestSignature{
		GroupID:  gid,
		FeeLimit: feeLimit,
		Sender:   sender.String(),
	}
	err := m.SetContent(content)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Route Implements Msg.
func (m MsgRequestSignature) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgRequestSignature) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgRequestSignature) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m *MsgRequestSignature) GetContent() Content {
	content, ok := m.Content.GetCachedValue().(Content)
	if !ok {
		return nil
	}
	return content
}

// GetSigners returns the expected signers for a MsgRequestSignature.
func (m MsgRequestSignature) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Sender)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgRequestSignature) ValidateBasic() error {
	// Validate sender address
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errors.Wrap(err, "sender")
	}

	return nil
}

func (m *MsgRequestSignature) SetContent(content Content) error {
	msg, ok := content.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return err
	}
	m.Content = any
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgRequestSignature) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var content Content
	return unpacker.UnpackAny(m.Content, &content)
}

// Route Implements Msg.
func (m MsgSubmitSignature) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgSubmitSignature) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgSubmitSignature) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgSubmitSignature.
func (m MsgSubmitSignature) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Member)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgSubmitSignature) ValidateBasic() error {
	// Validate member address
	_, err := sdk.AccAddressFromBech32(m.Member)
	if err != nil {
		return errors.Wrap(err, "member")
	}

	// Validate member signature
	if err = m.Signature.Validate(); err != nil {
		return errors.Wrap(err, "signature")
	}

	return nil
}

// Route Implements Msg.
func (m MsgActivate) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgActivate) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgActivate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgActivate.
func (m MsgActivate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Address)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgActivate) ValidateBasic() error {
	// Validate member address
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return errors.Wrap(err, "member")
	}

	return nil
}

// Route Implements Msg.
func (m MsgActive) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgActive) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgActive) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgActive.
func (m MsgActive) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Address)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgActive) ValidateBasic() error {
	// Validate member address
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return errors.Wrap(err, "member")
	}

	return nil
}

// NewMsgActivate creates a new MsgActivate instance
func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// Route Implements Msg.
func (m MsgUpdateParams) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgUpdateParams) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Authority)}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errors.Wrap(err, "invalid authority address")
	}

	if err := m.Params.Validate(); err != nil {
		return err
	}

	return nil
}
