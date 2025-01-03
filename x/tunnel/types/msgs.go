package types

import (
	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
)

var (
	_, _, _, _, _, _, _, _, _, _ sdk.Msg                       = &MsgCreateTunnel{}, &MsgUpdateRoute{}, &MsgUpdateSignalsAndInterval{}, &MsgWithdrawFeePayerFunds{}, &MsgActivate{}, &MsgDeactivate{}, &MsgTriggerTunnel{}, &MsgDepositToTunnel{}, &MsgWithdrawFromTunnel{}, &MsgUpdateParams{}
	_, _, _, _, _, _, _, _, _    sdk.HasValidateBasic          = &MsgCreateTunnel{}, &MsgUpdateRoute{}, &MsgUpdateSignalsAndInterval{}, &MsgActivate{}, &MsgDeactivate{}, &MsgTriggerTunnel{}, &MsgDepositToTunnel{}, &MsgWithdrawFromTunnel{}, &MsgUpdateParams{}
	_, _                         types.UnpackInterfacesMessage = &MsgCreateTunnel{}, &MsgUpdateRoute{}
)

// NewMsgCreateTunnel creates a new MsgCreateTunnel instance.
func NewMsgCreateTunnel(
	signalDeviations []SignalDeviation,
	interval uint64,
	route RouteI,
	initialDeposit sdk.Coins,
	creator string,
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
		InitialDeposit:   initialDeposit,
		Creator:          creator,
	}, nil
}

// NewMsgCreateTSSTunnel creates a new MsgCreateTunnel instance for TSS tunnel.
func NewMsgCreateTSSTunnel(
	signalDeviations []SignalDeviation,
	interval uint64,
	destinationChainID string,
	destinationContractAddress string,
	encoder feedstypes.Encoder,
	initialDeposit sdk.Coins,
	creator string,
) (*MsgCreateTunnel, error) {
	r := NewTSSRoute(destinationChainID, destinationContractAddress, encoder)
	m, err := NewMsgCreateTunnel(signalDeviations, interval, &r, initialDeposit, creator)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// NewMsgCreateIBCTunnel creates a new MsgCreateTunnel instance with IBC route type.
func NewMsgCreateIBCTunnel(
	signalDeviations []SignalDeviation,
	interval uint64,
	deposit sdk.Coins,
	creator string,
) (*MsgCreateTunnel, error) {
	r := NewIBCRoute("")
	m, err := NewMsgCreateTunnel(signalDeviations, interval, r, deposit, creator)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// NewMsgCreateRouterTunnel creates a new MsgCreateTunnel instance for Router tunnel.
func NewMsgCreateRouterTunnel(
	signalDeviations []SignalDeviation,
	interval uint64,
	fund sdk.Coin,
	bridgeContractAddress string,
	destChainID string,
	destContractAddress string,
	destGasLimit uint64,
	destGasPrice uint64,
	initialDeposit sdk.Coins,
	creator string,
) (*MsgCreateTunnel, error) {
	r := &RouterRoute{
		Fund:                  fund,
		BridgeContractAddress: bridgeContractAddress,
		DestChainID:           destChainID,
		DestContractAddress:   destContractAddress,
		DestGasLimit:          destGasLimit,
		DestGasPrice:          destGasPrice,
	}
	m, err := NewMsgCreateTunnel(signalDeviations, interval, r, initialDeposit, creator)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// GetRouteValue returns the route of the tunnel.
func (m MsgCreateTunnel) GetRouteValue() (RouteI, error) {
	r, ok := m.Route.GetCachedValue().(RouteI)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf("expected %T, got %T", (RouteI)(nil), m.Route.GetCachedValue())
	}

	return r, nil
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
	r, err := m.GetRouteValue()
	if err != nil {
		return err
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

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgCreateTunnel) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var route RouteI
	return unpacker.UnpackAny(m.Route, &route)
}

// NewMsgUpdateRoute creates a new MsgUpdateRoute instance.
func NewMsgUpdateRoute(
	tunnelID uint64,
	route RouteI,
	creator string,
) (*MsgUpdateRoute, error) {
	msg, ok := route.(proto.Message)
	if !ok {
		return nil, sdkerrors.ErrPackAny.Wrapf("cannot proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return &MsgUpdateRoute{
		TunnelID: tunnelID,
		Route:    any,
		Creator:  creator,
	}, nil
}

// NewMsgUpdateIBCRoute creates a new MsgUpdateRoute instance.
func NewMsgUpdateIBCRoute(
	tunnelID uint64,
	channelID string,
	creator string,
) (*MsgUpdateRoute, error) {
	return NewMsgUpdateRoute(tunnelID, NewIBCRoute(channelID), creator)
}

// GetRouteValue returns the route of the message.
func (m MsgUpdateRoute) GetRouteValue() (RouteI, error) {
	r, ok := m.Route.GetCachedValue().(RouteI)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf("expected %T, got %T", (RouteI)(nil), m.Route.GetCachedValue())
	}
	return r, nil
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateRoute) ValidateBasic() error {
	// creator address must be valid
	if _, err := sdk.AccAddressFromBech32(m.Creator); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	// route must be valid
	r, err := m.GetRouteValue()
	if err != nil {
		return err
	}

	if err := r.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgUpdateRoute) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var route RouteI
	return unpacker.UnpackAny(m.Route, &route)
}

// NewMsgUpdateSignalsAndInterval creates a new MsgUpdateSignalsAndInterval instance.
func NewMsgUpdateSignalsAndInterval(
	tunnelID uint64,
	signalDeviations []SignalDeviation,
	interval uint64,
	creator string,
) *MsgUpdateSignalsAndInterval {
	return &MsgUpdateSignalsAndInterval{
		TunnelID:         tunnelID,
		SignalDeviations: signalDeviations,
		Interval:         interval,
		Creator:          creator,
	}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateSignalsAndInterval) ValidateBasic() error {
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

// NewMsgWithdrawFeePayerFunds creates a new MsgWithdrawFeePayerFunds instance.
func NewMsgWithdrawFeePayerFunds(tunnelID uint64, amount sdk.Coins, creator string) *MsgWithdrawFeePayerFunds {
	return &MsgWithdrawFeePayerFunds{
		TunnelID: tunnelID,
		Amount:   amount,
		Creator:  creator,
	}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgWithdrawFeePayerFunds) ValidateBasic() error {
	// creator address must be valid
	if _, err := sdk.AccAddressFromBech32(m.Creator); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	// amount must be valid
	if !m.Amount.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrapf("invalid funds: %s", m.Amount)
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

// NewMsgDepositToTunnel creates a new MsgDepositToTunnel instance.
func NewMsgDepositToTunnel(
	tunnelID uint64,
	amount sdk.Coins,
	depositor string,
) *MsgDepositToTunnel {
	return &MsgDepositToTunnel{
		TunnelID:  tunnelID,
		Amount:    amount,
		Depositor: depositor,
	}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgDepositToTunnel) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Depositor); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	if !m.Amount.IsValid() || !m.Amount.IsAllPositive() {
		return sdkerrors.ErrInvalidCoins.Wrapf("invalid amount: %s", m.Amount)
	}

	return nil
}

// NewMsgWithdrawFromTunnel creates a new MsgWithdrawFromTunnel instance.
func NewMsgWithdrawFromTunnel(
	tunnelID uint64,
	amount sdk.Coins,
	withdrawer string,
) *MsgWithdrawFromTunnel {
	return &MsgWithdrawFromTunnel{
		TunnelID:   tunnelID,
		Amount:     amount,
		Withdrawer: withdrawer,
	}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgWithdrawFromTunnel) ValidateBasic() error {
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
