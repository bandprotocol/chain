package types

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	proto "github.com/cosmos/gogoproto/proto"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

var (
	_ sdk.Msg = &MsgTransitionGroup{}
	_ sdk.Msg = &MsgForceReplaceGroup{}
	_ sdk.Msg = &MsgRequestSignature{}
	_ sdk.Msg = &MsgActivate{}
	_ sdk.Msg = &MsgHeartbeat{}
	_ sdk.Msg = &MsgUpdateParams{}

	_ types.UnpackInterfacesMessage = &MsgRequestSignature{}
)

// NewMsgTransitionGroup creates a new MsgTransitionGroup instance.
func NewMsgTransitionGroup(
	members []string,
	threshold uint64,
	execTime time.Time,
	authority string,
) *MsgTransitionGroup {
	return &MsgTransitionGroup{
		Members:   members,
		Threshold: threshold,
		Authority: authority,
		ExecTime:  execTime,
	}
}

// Type returns message type name.
func (m MsgTransitionGroup) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgTransitionGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgTransitionGroup.
func (m MsgTransitionGroup) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Authority)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgTransitionGroup) ValidateBasic() error {
	// Validate members address
	for _, member := range m.Members {
		if _, err := sdk.AccAddressFromBech32(member); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid member address: %s", err)
		}
	}

	// Check duplicate member
	if tsstypes.DuplicateInArray(m.Members) {
		return ErrMemberDuplicate
	}

	// Validate sender address
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	// Validate threshold must be less than or equal to members but more than zero
	if m.Threshold > uint64(len(m.Members)) || m.Threshold <= 0 {
		return ErrInvalidSigningThreshold.Wrapf(
			"threshold must be less than or equal to the members but more than zero",
		)
	}

	return nil
}

// NewMsgForceReplaceGroup creates a new NewMsgForceReplaceGroup instance.
func NewMsgForceReplaceGroup(
	incomingGroupID tss.GroupID,
	execTime time.Time,
	authority string,
) *MsgForceReplaceGroup {
	return &MsgForceReplaceGroup{
		IncomingGroupID: incomingGroupID,
		ExecTime:        execTime,
		Authority:       authority,
	}
}

// Type returns message type name.
func (m MsgForceReplaceGroup) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgForceReplaceGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a NewMsgForceReplaceGroup.
func (m MsgForceReplaceGroup) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Authority)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgForceReplaceGroup) ValidateBasic() error {
	// Validate sender address
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	if m.IncomingGroupID == 0 {
		return ErrInvalidIncomingGroup.Wrap("incoming group ID must not be zero")
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
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
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
func NewMsgActivate(sender string, groupID tss.GroupID) *MsgActivate {
	return &MsgActivate{
		Sender:  sender,
		GroupID: groupID,
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
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Sender)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgActivate) ValidateBasic() error {
	// Validate member address
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	return nil
}

// NewMsgHeartbeat creates a new MsgHeartbeat instance.
func NewMsgHeartbeat(sender string, groupID tss.GroupID) *MsgHeartbeat {
	return &MsgHeartbeat{
		Sender:  sender,
		GroupID: groupID,
	}
}

// Type returns message type name.
func (m MsgHeartbeat) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgHeartbeat) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgHeartbeat.
func (m MsgHeartbeat) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Sender)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgHeartbeat) ValidateBasic() error {
	// Validate member address
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
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
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	if err := m.Params.Validate(); err != nil {
		return err
	}

	return nil
}
