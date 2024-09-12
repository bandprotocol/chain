package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
)

var (
	_, _, _, _, _, _, _ sdk.Msg                       = &MsgCreateTunnel{}, &MsgEditTunnel{}, &MsgActivateTunnel{}, &MsgDeactivateTunnel{}, &MsgManualTriggerTunnel{}, &MsgDeposit{}, &MsgUpdateParams{}
	_                   types.UnpackInterfacesMessage = &MsgCreateTunnel{}
)

// NewMsgUpdateParams creates a new MsgUpdateParams instance.
func NewMsgUpdateParams(
	authority string,
	params Params,
) *MsgUpdateParams {
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
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Authority)}
}

// ValidateBasic does a check on the provided data.
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	if err := m.Params.Validate(); err != nil {
		return err
	}

	return nil
}

func NewMsgCreateTunnel(
	signalInfos []SignalInfo,
	interval uint64,
	route RouteI,
	encoder Encoder,
	initialDeposit sdk.Coins,
	creator sdk.AccAddress,
) (*MsgCreateTunnel, error) {
	msg, ok := route.(proto.Message)
	if !ok {
		return nil, sdkerrors.ErrPackAny.Wrapf("cannot proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return &MsgCreateTunnel{
		SignalInfos:    signalInfos,
		Interval:       interval,
		Route:          any,
		Encoder:        encoder,
		InitialDeposit: initialDeposit,
		Creator:        creator.String(),
	}, nil
}

// NewMsgCreateTunnel creates a new MsgCreateTunnel instance.
func NewMsgCreateTSSTunnel(
	signalInfos []SignalInfo,
	interval uint64,
	encoder Encoder,
	destinationChainID string,
	destinationContractAddress string,
	initialDeposit sdk.Coins,
	creator sdk.AccAddress,
) (*MsgCreateTunnel, error) {
	r := &TSSRoute{
		DestinationChainID:         destinationChainID,
		DestinationContractAddress: destinationContractAddress,
	}
	m, err := NewMsgCreateTunnel(signalInfos, interval, r, encoder, initialDeposit, creator)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// NewMsgCreateTunnel creates a new MsgCreateTunnel instance.
func NewMsgCreateAxelarTunnel(
	signalInfos []SignalInfo,
	interval uint64,
	encoder Encoder,
	destinationChainID string,
	destinationContractAddress string,
	deposit sdk.Coins,
	creator sdk.AccAddress,
) (*MsgCreateTunnel, error) {
	r := &TSSRoute{
		DestinationChainID:         destinationChainID,
		DestinationContractAddress: destinationContractAddress,
	}
	m, err := NewMsgCreateTunnel(signalInfos, interval, r, encoder, deposit, creator)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// Type Implements Msg.
func (m MsgCreateTunnel) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgCreateTunnel) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgCreateTunnel) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Creator)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgCreateTunnel) ValidateBasic() error {
	// creator address must be valid
	if _, err := sdk.AccAddressFromBech32(m.Creator); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	// signalInfos must not be empty
	if len(m.SignalInfos) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("signal infos cannot be empty")
	}

	// route must be valid
	r, ok := m.Route.GetCachedValue().(RouteI)
	if !ok {
		return sdkerrors.ErrPackAny.Wrapf("cannot unpack route")
	}
	if err := r.ValidateBasic(); err != nil {
		return err
	}

	// initialDeposit deposit must be positive
	if !m.InitialDeposit.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrapf("invalid deposit: %s", m.InitialDeposit)
	}

	// signalIDs must be unique
	signalIDMap := make(map[string]bool)
	for _, signalInfo := range m.SignalInfos {
		if _, ok := signalIDMap[signalInfo.SignalID]; ok {
			return sdkerrors.ErrInvalidRequest.Wrapf("duplicate signal ID: %s", signalInfo.SignalID)
		}
		signalIDMap[signalInfo.SignalID] = true
	}

	return nil
}

// SetRoute sets the route for the message.
func (m *MsgCreateTunnel) SetTunnelRoute(route RouteI) error {
	msg, ok := route.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return err
	}
	m.Route = any

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgCreateTunnel) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var route RouteI
	return unpacker.UnpackAny(m.Route, &route)
}

// GetRoute returns the route of the message.
func (m MsgCreateTunnel) GetTunnelRoute() RouteI {
	route, ok := m.Route.GetCachedValue().(RouteI)
	if !ok {
		return nil
	}

	return route
}

// NewMsgEditTunnel creates a new MsgEditTunnel instance.
func NewMsgEditTunnel(
	tunnelID uint64,
	signalInfos []SignalInfo,
	interval uint64,
	creator string,
) *MsgEditTunnel {
	return &MsgEditTunnel{
		TunnelID:    tunnelID,
		SignalInfos: signalInfos,
		Interval:    interval,
		Creator:     creator,
	}
}

// Route Implements Msg.
func (m MsgEditTunnel) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgEditTunnel) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgEditTunnel) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Creator)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgEditTunnel) ValidateBasic() error {
	// creator address must be valid
	if _, err := sdk.AccAddressFromBech32(m.Creator); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	// signalIDs must be unique
	signalIDMap := make(map[string]bool)
	for _, signalInfo := range m.SignalInfos {
		if _, ok := signalIDMap[signalInfo.SignalID]; ok {
			return sdkerrors.ErrInvalidRequest.Wrapf("duplicate signal ID: %s", signalInfo.SignalID)
		}
		signalIDMap[signalInfo.SignalID] = true
	}

	return nil
}

// NewMsgActivateTunnel creates a new MsgActivateTunnel instance.
func NewMsgActivateTunnel(
	tunnelID uint64,
	creator string,
) *MsgActivateTunnel {
	return &MsgActivateTunnel{
		TunnelID: tunnelID,
		Creator:  creator,
	}
}

// Route Implements Msg.
func (m MsgActivateTunnel) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgActivateTunnel) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgActivateTunnel) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Creator)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgActivateTunnel) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Creator); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	return nil
}

// NewMsgDeactivateTunnel creates a new MsgDeactivateTunnel instance.
func NewMsgDeactivateTunnel(
	tunnelID uint64,
	creator string,
) *MsgDeactivateTunnel {
	return &MsgDeactivateTunnel{
		TunnelID: tunnelID,
		Creator:  creator,
	}
}

// Route Implements Msg.
func (m MsgDeactivateTunnel) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgDeactivateTunnel) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgDeactivateTunnel) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Creator)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgDeactivateTunnel) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Creator); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	return nil
}

// NewMsgManualTriggerTunnel creates a new MsgManualTriggerTunnel instance.
func NewMsgManualTriggerTunnel(
	tunnelID uint64,
	creator string,
) *MsgManualTriggerTunnel {
	return &MsgManualTriggerTunnel{
		TunnelID: tunnelID,
		Creator:  creator,
	}
}

// Route Implements Msg.
func (m MsgManualTriggerTunnel) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgManualTriggerTunnel) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgManualTriggerTunnel) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Creator)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgManualTriggerTunnel) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Creator); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	return nil
}

// NewMsgDeposit creates a new MsgDeposit instance.
func NewMsgDeposit(
	tunnelID uint64,
	amount sdk.Coins,
	depositor string,
) *MsgDeposit {
	return &MsgDeposit{
		TunnelId:  tunnelID,
		Amount:    amount,
		Depositor: depositor,
	}
}

// Route Implements Msg.
func (m MsgDeposit) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgDeposit) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgDeposit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Depositor)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgDeposit) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Depositor); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	if !m.Amount.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrapf("invalid amount: %s", m.Amount)
	}

	return nil
}
