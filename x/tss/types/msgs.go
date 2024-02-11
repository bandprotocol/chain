package types

import (
	"fmt"
	"time"

	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	proto "github.com/gogo/protobuf/proto"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

var (
	_, _, _, _, _, _    sdk.Msg                       = &MsgCreateGroup{}, &MsgSubmitDKGRound1{}, &MsgSubmitDKGRound2{}, &MsgComplain{}, &MsgConfirm{}, &MsgSubmitDEs{}
	_, _, _, _, _, _, _ sdk.Msg                       = &MsgRequestSignature{}, &MsgSubmitSignature{}, &MsgActivate{}, &MsgHealthCheck{}, &MsgReplaceGroup{}, &MsgUpdateParams{}, &MsgUpdateGroupFee{}
	_                   types.UnpackInterfacesMessage = &MsgRequestSignature{}
)

// NewMsgCreateGroup creates a new MsgCreateGroup instance.
func NewMsgCreateGroup(members []string, threshold uint64, fee sdk.Coins, authority string) *MsgCreateGroup {
	return &MsgCreateGroup{
		Members:   members,
		Threshold: threshold,
		Fee:       fee,
		Authority: authority,
	}
}

// Type returns message type name.
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
				fmt.Sprintf("member: %s", member),
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

	// Validate fee
	if !m.Fee.IsValid() {
		return errors.Wrap(sdkerrors.ErrInvalidCoins, m.Fee.String())
	}

	return nil
}

// NewMsgReplaceGroup creates a new MsgReplaceGroup instance.
func NewMsgReplaceGroup(
	currentGroupID tss.GroupID,
	newGroupID tss.GroupID,
	execTime time.Time,
	authority string,
) *MsgReplaceGroup {
	return &MsgReplaceGroup{
		CurrentGroupID: currentGroupID,
		NewGroupID:     newGroupID,
		ExecTime:       execTime,
		Authority:      authority,
	}
}

// Type returns message type name.
func (m MsgReplaceGroup) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgReplaceGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgReplaceGroup.
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

// NewMsgSubmitDKGRound1 creates a new MsgSubmitDKGRound1 instance.
func NewMsgSubmitDKGRound1(groupID tss.GroupID, round1Info Round1Info, address string) *MsgSubmitDKGRound1 {
	return &MsgSubmitDKGRound1{
		GroupID:    groupID,
		Round1Info: round1Info,
		Address:    address,
	}
}

// Type returns message type name.
func (m MsgSubmitDKGRound1) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgSubmitDKGRound1) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgSubmitDKGRound1.
func (m MsgSubmitDKGRound1) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Address)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgSubmitDKGRound1) ValidateBasic() error {
	// Validate member address
	_, err := sdk.AccAddressFromBech32(m.Address)
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
	if err := m.Round1Info.A0Signature.Validate(); err != nil {
		return errors.Wrap(err, "a0 sig")
	}

	// Validate one time signature
	if err := m.Round1Info.OneTimeSignature.Validate(); err != nil {
		return errors.Wrap(err, "one time sig")
	}

	return nil
}

// NewMsgSubmitDKGRound2 creates a new MsgSubmitDKGRound2 instance.
func NewMsgSubmitDKGRound2(groupID tss.GroupID, round2Info Round2Info, address string) *MsgSubmitDKGRound2 {
	return &MsgSubmitDKGRound2{
		GroupID:    groupID,
		Round2Info: round2Info,
		Address:    address,
	}
}

// Type returns message type name.
func (m MsgSubmitDKGRound2) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgSubmitDKGRound2) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgSubmitDKGRound2.
func (m MsgSubmitDKGRound2) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Address)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgSubmitDKGRound2) ValidateBasic() error {
	// Validate member address
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return errors.Wrap(err, "member")
	}

	// Validate encrypted secret shares
	for i, ess := range m.Round2Info.EncryptedSecretShares {
		if err := ess.Validate(); err != nil {
			return errors.Wrap(err, fmt.Sprintf("encrypted secret shares at index: %d", i))
		}
	}

	return nil
}

// NewMsgComplain creates a new MsgComplain instance.
func NewMsgComplain(groupID tss.GroupID, complaints []Complaint, address string) *MsgComplain {
	return &MsgComplain{
		GroupID:    groupID,
		Complaints: complaints,
		Address:    address,
	}
}

// Type returns message type name.
func (m MsgComplain) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgComplain) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgComplain.
func (m MsgComplain) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Address)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgComplain) ValidateBasic() error {
	// Validate member address
	_, err := sdk.AccAddressFromBech32(m.Address)
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

// NewMsgConfirm creates a new MsgConfirm instance.
func NewMsgConfirm(
	groupID tss.GroupID,
	memberID tss.MemberID,
	ownPubKeySig tss.Signature,
	address string,
) *MsgConfirm {
	return &MsgConfirm{
		GroupID:      groupID,
		MemberID:     memberID,
		OwnPubKeySig: ownPubKeySig,
		Address:      address,
	}
}

// Type returns message type name.
func (m MsgConfirm) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgConfirm) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgConfirm.
func (m MsgConfirm) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Address)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgConfirm) ValidateBasic() error {
	// Validate member address
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return errors.Wrap(err, "member")
	}

	// Validate own pub key sig
	if err = m.OwnPubKeySig.Validate(); err != nil {
		return errors.Wrap(err, "own pub key sig")
	}

	return nil
}

// NewMsgSubmitDEs creates a new MsgSubmitDEs instance.
func NewMsgSubmitDEs(des []DE, address string) *MsgSubmitDEs {
	return &MsgSubmitDEs{
		DEs:     des,
		Address: address,
	}
}

// Type returns message type name.
func (m MsgSubmitDEs) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgSubmitDEs) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgSubmitDEs.
func (m MsgSubmitDEs) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Address)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgSubmitDEs) ValidateBasic() error {
	// Validate member address
	_, err := sdk.AccAddressFromBech32(m.Address)
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

// Type returns message type name.
func (m MsgRequestSignature) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgRequestSignature) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetContent returns the content of MsgRequestSignature.
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

// SetContent sets the content for MsgRequestSignature.
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

// NewMsgSubmitSignature creates a new MsgSubmitSignature instance.
func NewMsgSubmitSignature(
	signingID tss.SigningID,
	memberID tss.MemberID,
	signature tss.Signature,
	address string,
) *MsgSubmitSignature {
	return &MsgSubmitSignature{
		SigningID: signingID,
		MemberID:  memberID,
		Signature: signature,
		Address:   address,
	}
}

// Type returns message type name.
func (m MsgSubmitSignature) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgSubmitSignature) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgSubmitSignature.
func (m MsgSubmitSignature) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Address)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgSubmitSignature) ValidateBasic() error {
	// Validate member address
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return errors.Wrap(err, "member")
	}

	// Validate member signature
	if err = m.Signature.Validate(); err != nil {
		return errors.Wrap(err, "signature")
	}

	return nil
}

// NewMsgActivate creates a new MsgActivate instance.
func NewMsgActivate(address string) *MsgActivate {
	return &MsgActivate{
		Address: address,
	}
}

// Type returns message type name.
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

// NewMsgHealthCheck creates a new MsgHealthCheck instance.
func NewMsgHealthCheck(address string) *MsgHealthCheck {
	return &MsgHealthCheck{
		Address: address,
	}
}

// Type returns message type name.
func (m MsgHealthCheck) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgHealthCheck) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgHealthCheck.
func (m MsgHealthCheck) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Address)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgHealthCheck) ValidateBasic() error {
	// Validate member address
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return errors.Wrap(err, "member")
	}

	return nil
}

// NewMsgUpdateGroupFee creates a new MsgUpdateGroupFee instance.
func NewMsgUpdateGroupFee(groupID tss.GroupID, fee sdk.Coins, authority string) *MsgUpdateGroupFee {
	return &MsgUpdateGroupFee{
		GroupID:   groupID,
		Fee:       fee,
		Authority: authority,
	}
}

// Type returns message type name.
func (m MsgUpdateGroupFee) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupFee) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupFee.
func (m MsgUpdateGroupFee) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Authority)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateGroupFee) ValidateBasic() error {
	// Validate sender address
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf("sender: %s", m.Authority),
		)
	}

	// Validate fee
	if !m.Fee.IsValid() {
		return errors.Wrap(sdkerrors.ErrInvalidCoins, m.Fee.String())
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
