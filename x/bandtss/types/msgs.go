package types

import (
	"fmt"
	"time"

	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/cosmos/gogoproto/proto"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

var (
	_ sdk.Msg = &MsgCreateGroup{}
	_ sdk.Msg = &MsgReplaceGroup{}
	_ sdk.Msg = &MsgRequestSignature{}
	_ sdk.Msg = &MsgActivate{}
	_ sdk.Msg = &MsgHealthCheck{}
	_ sdk.Msg = &MsgUpdateParams{}

	_ types.UnpackInterfacesMessage = &MsgRequestSignature{}
)

// NewMsgCreateGroup creates a new MsgCreateGroup instance.
func NewMsgCreateGroup(members []string, threshold uint64, fee sdk.Coins, authority string) *MsgCreateGroup {
	return &MsgCreateGroup{
		Members:   members,
		Threshold: threshold,
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
		NewGroupID: newGroupID,
		ExecTime:   execTime,
		Authority:  authority,
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

// NewMsgRequestSignature creates a new MsgRequestSignature.
func NewMsgRequestSignature(
	content tsstypes.Content,
	feeLimit sdk.Coins,
	sender sdk.AccAddress,
) (*MsgRequestSignature, error) {
	m := &MsgRequestSignature{
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
func (m *MsgRequestSignature) GetContent() tsstypes.Content {
	content, ok := m.Content.GetCachedValue().(tsstypes.Content)
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
func (m *MsgRequestSignature) SetContent(content tsstypes.Content) error {
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
	var content tsstypes.Content
	return unpacker.UnpackAny(m.Content, &content)
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
