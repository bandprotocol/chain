package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
)

var (
	_, _, _, _, _, _, _, _ sdk.Msg                       = &MsgCreateTunnel{}, &MsgEditTunnel{}, &MsgActivate{}, &MsgDeactivate{}, &MsgTriggerTunnel{}, &MsgDepositTunnel{}, &MsgWithdrawDepositTunnel{}, &MsgUpdateParams{}
	_                      types.UnpackInterfacesMessage = &MsgCreateTunnel{}
)

// NewMsgCreateTunnel creates a new MsgCreateTunnel instance.
func NewMsgCreateTunnel(
	signalDeviations []SignalDeviation,
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
		SignalDeviations: signalDeviations,
		Interval:         interval,
		Route:            any,
		Encoder:          encoder,
		InitialDeposit:   initialDeposit,
		Creator:          creator.String(),
	}, nil
}

// NewMsgCreateTSSTunnel creates a new MsgCreateTunnel instance for TSS tunnel.
func NewMsgCreateTSSTunnel(
	signalDeviations []SignalDeviation,
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
	m, err := NewMsgCreateTunnel(signalDeviations, interval, r, encoder, initialDeposit, creator)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// NewMsgCreateAxelarTunnel creates a new MsgCreateTunnel instance for Axelar tunnel.
func NewMsgCreateAxelarTunnel(
	signalDeviations []SignalDeviation,
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
	m, err := NewMsgCreateTunnel(signalDeviations, interval, r, encoder, deposit, creator)
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

	// signal deviations cannot be empty
	if len(m.SignalDeviations) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("signal deviations cannot be empty")
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
	for _, signalDeviation := range m.SignalDeviations {
		if _, ok := signalIDMap[signalDeviation.SignalID]; ok {
			return sdkerrors.ErrInvalidRequest.Wrapf("duplicate signal ID: %s", signalDeviation.SignalID)
		}

		signalIDMap[signalDeviation.SignalID] = true
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
	signalDeviations []SignalDeviation,
	interval uint64,
	creator string,
) *MsgEditTunnel {
	return &MsgEditTunnel{
		TunnelID:         tunnelID,
		SignalDeviations: signalDeviations,
		Interval:         interval,
		Creator:          creator,
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
	for _, signalDeviation := range m.SignalDeviations {
		if _, ok := signalIDMap[signalDeviation.SignalID]; ok {
			return sdkerrors.ErrInvalidRequest.Wrapf("duplicate signal ID: %s", signalDeviation.SignalID)
		}
		signalIDMap[signalDeviation.SignalID] = true
	}

	err := validateUniqueSignalIDs(m.SignalDeviations)
	if err != nil {
		return err
	}

	return nil
}

// NewMsgActivate creates a new MsgActivate instance.
func NewMsgActivate(
	tunnelID uint64,
	creator string,
) *MsgActivate {
	return &MsgActivate{
		TunnelID: tunnelID,
		Creator:  creator,
	}
}

// Route Implements Msg.
func (m MsgActivate) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgActivate) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgActivate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Creator)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgActivate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Creator); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	return nil
}

// NewMsgDeactivate creates a new MsgDeactivate instance.
func NewMsgDeactivate(
	tunnelID uint64,
	creator string,
) *MsgDeactivate {
	return &MsgDeactivate{
		TunnelID: tunnelID,
		Creator:  creator,
	}
}

// Route Implements Msg.
func (m MsgDeactivate) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgDeactivate) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgDeactivate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Creator)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgDeactivate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Creator); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	return nil
}

// NewMsgTriggerTunnel creates a new MsgTriggerTunnel instance.
func NewMsgTriggerTunnel(
	tunnelID uint64,
	creator string,
) *MsgTriggerTunnel {
	return &MsgTriggerTunnel{
		TunnelID: tunnelID,
		Creator:  creator,
	}
}

// Route Implements Msg.
func (m MsgTriggerTunnel) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgTriggerTunnel) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgTriggerTunnel) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Creator)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgTriggerTunnel) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Creator); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	return nil
}

// NewMsgDepositTunnel creates a new MsgDeposit instance.
func NewMsgDepositTunnel(
	tunnelID uint64,
	amount sdk.Coins,
	depositor string,
) *MsgDepositTunnel {
	return &MsgDepositTunnel{
		TunnelID:  tunnelID,
		Amount:    amount,
		Depositor: depositor,
	}
}

// Route Implements Msg.
func (m MsgDepositTunnel) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgDepositTunnel) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgDepositTunnel) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Depositor)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgDepositTunnel) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Depositor); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	if !m.Amount.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrapf("invalid amount: %s", m.Amount)
	}

	return nil
}

// NewMsgWithdrawDepositTunnel creates a new MsgWithdraw instance.
func NewMsgWithdrawDepositTunnel(
	tunnelID uint64,
	amount sdk.Coins,
	withdrawer string,
) *MsgWithdrawDepositTunnel {
	return &MsgWithdrawDepositTunnel{
		TunnelID:   tunnelID,
		Amount:     amount,
		Withdrawer: withdrawer,
	}
}

// Route Implements Msg.
func (m MsgWithdrawDepositTunnel) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgWithdrawDepositTunnel) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgWithdrawDepositTunnel) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Withdrawer)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgWithdrawDepositTunnel) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Withdrawer); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	if !m.Amount.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrapf("invalid amount: %s", m.Amount)
	}

	return nil
}

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

// validateUniqueSignalIDs checks if the SignalIDs in the given slice are unique
func validateUniqueSignalIDs(signalDeviations []SignalDeviation) error {
	signalIDMap := make(map[string]bool)
	for _, signalDeviation := range signalDeviations {
		if _, ok := signalIDMap[signalDeviation.SignalID]; ok {
			return sdkerrors.ErrInvalidRequest.Wrapf("duplicate signal ID: %s", signalDeviation.SignalID)
		}
		signalIDMap[signalDeviation.SignalID] = true
	}
	return nil
}
