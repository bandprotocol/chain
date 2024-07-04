package types

import (
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	proto "github.com/cosmos/gogoproto/proto"

	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
)

var _ sdk.Msg = &MsgUpdateParams{}

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
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
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
	feedType feedstypes.FeedType,
	route Route,
	deposit sdk.Coins,
	creator sdk.AccAddress,
) (*MsgCreateTunnel, error) {
	msg, ok := route.(proto.Message)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrPackAny, "cannot proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return &MsgCreateTunnel{
		SignalInfos: signalInfos,
		Route:       any,
		FeedType:    feedType,
		Deposit:     deposit,
		Creator:     creator.String(),
	}, nil
}

// NewMsgCreateTunnel creates a new MsgCreateTunnel instance.
func NewMsgCreateTSSTunnel(
	signalInfos []SignalInfo,
	feedType feedstypes.FeedType,
	destinationChainID string,
	destinationContractAddress string,
	deposit sdk.Coins,
	creator sdk.AccAddress,
) (*MsgCreateTunnel, error) {
	m := &MsgCreateTunnel{
		SignalInfos: signalInfos,
		FeedType:    feedType,
		Deposit:     deposit,
		Creator:     creator.String(),
	}

	r := &TSSRoute{
		DestinationChainID:         destinationChainID,
		DestinationContractAddress: destinationContractAddress,
	}

	fmt.Printf("tssroute: %+v\n", r)

	err := m.SetTunnelRoute(r)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// NewMsgCreateTunnel creates a new MsgCreateTunnel instance.
func NewMsgCreateAxelarTunnel(
	signalInfos []SignalInfo,
	feedType feedstypes.FeedType,
	destinationChainID string,
	destinationContractAddress string,
	deposit sdk.Coins,
	creator sdk.AccAddress,
) (*MsgCreateTunnel, error) {
	m := &MsgCreateTunnel{
		SignalInfos: signalInfos,
		FeedType:    feedType,
		Deposit:     deposit,
		Creator:     creator.String(),
	}

	err := m.SetTunnelRoute(&AxelarRoute{
		DestinationChainID:         destinationChainID,
		DestinationContractAddress: destinationContractAddress,
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Type Implements Msg.
func (m MsgCreateTunnel) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgCreateTunnel) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgCreateTunnel) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Creator)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgCreateTunnel) ValidateBasic() error {
	return nil
}

// SetRoute sets the route for the message.
func (m *MsgCreateTunnel) SetTunnelRoute(route Route) error {
	msg, ok := route.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return err
	}
	m.Route = any

	fmt.Printf("set route: %+v\n", m.Route.GetCachedValue())

	return nil
}

// GetRoute returns the route of the message.
func (m MsgCreateTunnel) GetTunnelRoute() Route {
	route, ok := m.Route.GetCachedValue().(Route)
	if !ok {
		return nil
	}
	return route
}
