package types

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_, _, _, _, _, _, _, _ sdk.Msg                       = &MsgCreateTunnel{}, &MsgEditTunnel{}, &MsgActivate{}, &MsgDeactivate{}, &MsgTriggerTunnel{}, &MsgDepositTunnel{}, &MsgWithdrawTunnel{}, &MsgUpdateParams{}
	_, _, _, _, _, _, _, _ sdk.HasValidateBasic          = &MsgCreateTunnel{}, &MsgEditTunnel{}, &MsgActivate{}, &MsgDeactivate{}, &MsgTriggerTunnel{}, &MsgDepositTunnel{}, &MsgWithdrawTunnel{}, &MsgUpdateParams{}
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
	destinationChainID string,
	destinationContractAddress string,
	encoder Encoder,
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

// NewMsgCreateTunnel creates a new MsgCreateTunnel instance.
func NewMsgCreateIBCTunnel(
	signalDeviations []SignalDeviation,
	interval uint64,
	channelID string,
	encoder Encoder,
	deposit sdk.Coins,
	creator sdk.AccAddress,
) (*MsgCreateTunnel, error) {
	r := NewIBCRoute(channelID)
	m, err := NewMsgCreateTunnel(signalDeviations, interval, r, encoder, deposit, creator)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// NewMsgCreateRouterTunnel creates a new MsgCreateTunnel instance for Router tunnel.
func NewMsgCreateRouterTunnel(
	signalDeviations []SignalDeviation,
	interval uint64,
	channelID string,
	fund sdk.Coin,
	bridgeContractAddress string,
	destChainID string,
	destContractAddress string,
	destGasLimit uint64,
	destGasPrice uint64,
	encoder Encoder,
	initialDeposit sdk.Coins,
	creator sdk.AccAddress,
) (*MsgCreateTunnel, error) {
	r := &RouterRoute{
		ChannelID:             channelID,
		Fund:                  fund,
		BridgeContractAddress: bridgeContractAddress,
		DestChainID:           destChainID,
		DestContractAddress:   destContractAddress,
		DestGasLimit:          destGasLimit,
		DestGasPrice:          destGasPrice,
	}
	m, err := NewMsgCreateTunnel(signalDeviations, interval, r, encoder, initialDeposit, creator)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// Type Implements Msg.
func (m MsgCreateTunnel) Type() string { return sdk.MsgTypeURL(&m) }

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
	// signal deviations cannot duplicate
	if err := validateUniqueSignalIDs(m.SignalDeviations); err != nil {
		return err
	}

	// route must be valid
	r, ok := m.Route.GetCachedValue().(RouteI)
	if !ok {
		return sdkerrors.ErrPackAny.Wrapf("cannot unpack route")
	}
	if err := r.ValidateBasic(); err != nil {
		return err
	}

	// initial deposit must be valid
	if !m.InitialDeposit.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrapf("invalid initial deposit: %s", m.InitialDeposit)
	}

	return nil
}

// SetTunnelRoute sets the route of the tunnel.
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

// GetTunnelRoute returns the route of the tunnel.
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

// ValidateBasic does a sanity check on the provided data
func (m MsgEditTunnel) ValidateBasic() error {
	// creator address must be valid
	if _, err := sdk.AccAddressFromBech32(m.Creator); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	// signal deviations cannot be empty
	if len(m.SignalDeviations) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("signal deviations cannot be empty")
	}
	// signal deviations cannot duplicate
	if err := validateUniqueSignalIDs(m.SignalDeviations); err != nil {
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

// ValidateBasic does a sanity check on the provided data
func (m MsgDepositTunnel) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Depositor); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	if !m.Amount.IsValid() || !m.Amount.IsAllPositive() {
		return sdkerrors.ErrInvalidCoins.Wrapf("invalid amount: %s", m.Amount)
	}

	return nil
}

// NewMsgWithdrawTunnel creates a new MsgWithdrawTunnel instance.
func NewMsgWithdrawTunnel(
	tunnelID uint64,
	amount sdk.Coins,
	withdrawer string,
) *MsgWithdrawTunnel {
	return &MsgWithdrawTunnel{
		TunnelID:   tunnelID,
		Amount:     amount,
		Withdrawer: withdrawer,
	}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgWithdrawTunnel) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Withdrawer); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	if !m.Amount.IsValid() || !m.Amount.IsAllPositive() {
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
	for _, sd := range signalDeviations {
		if _, found := signalIDMap[sd.SignalID]; found {
			return sdkerrors.ErrInvalidRequest.Wrapf("duplicate signal ID: %s", sd.SignalID)
		}
		signalIDMap[sd.SignalID] = true
	}
	return nil
}
