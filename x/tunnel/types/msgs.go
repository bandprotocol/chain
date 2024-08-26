package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	proto "github.com/cosmos/gogoproto/proto"

	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
)

var (
	_, _, _, _ sdk.Msg                       = &MsgUpdateParams{}, &MsgCreateTunnel{}, &MsgActivateTunnel{}, &MsgManualTriggerTunnel{}
	_          types.UnpackInterfacesMessage = &MsgCreateTunnel{}
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
	feedType feedstypes.FeedType,
	deposit sdk.Coins,
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
		SignalInfos: signalInfos,
		Interval:    interval,
		Route:       any,
		FeedType:    feedType,
		Deposit:     deposit,
		Creator:     creator.String(),
	}, nil
}

// NewMsgCreateTunnel creates a new MsgCreateTunnel instance.
func NewMsgCreateTSSTunnel(
	signalInfos []SignalInfo,
	interval uint64,
	feedType feedstypes.FeedType,
	destinationChainID string,
	destinationContractAddress string,
	deposit sdk.Coins,
	creator sdk.AccAddress,
) (*MsgCreateTunnel, error) {
	r := &TSSRoute{
		DestinationChainID:         destinationChainID,
		DestinationContractAddress: destinationContractAddress,
	}
	m, err := NewMsgCreateTunnel(signalInfos, interval, r, feedType, deposit, creator)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// NewMsgCreateTunnel creates a new MsgCreateTunnel instance.
func NewMsgCreateAxelarTunnel(
	signalInfos []SignalInfo,
	interval uint64,
	feedType feedstypes.FeedType,
	destinationChainID string,
	destinationContractAddress string,
	deposit sdk.Coins,
	creator sdk.AccAddress,
) (*MsgCreateTunnel, error) {
	r := &TSSRoute{
		DestinationChainID:         destinationChainID,
		DestinationContractAddress: destinationContractAddress,
	}
	m, err := NewMsgCreateTunnel(signalInfos, interval, r, feedType, deposit, creator)
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
	r, ok := m.Route.GetCachedValue().(RouteI)
	if !ok {
		return sdkerrors.ErrPackAny.Wrapf("cannot unpack route")
	}
	if err := r.ValidateBasic(); err != nil {
		return err
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

func NewMsgManualTriggerTunnel(
	tunnelID uint64,
	creator string,
) *MsgActivateTunnel {
	return &MsgActivateTunnel{
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
