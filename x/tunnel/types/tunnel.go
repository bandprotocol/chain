package types

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.UnpackInterfacesMessage = Tunnel{}

// NewTunnel creates a new Tunnel instance.
func NewTunnel(
	id uint64,
	sequence uint64,
	route RouteI,
	encoder Encoder,
	feePayer string,
	signalDeviations []SignalDeviation,
	interval uint64,
	totalDeposit []sdk.Coin,
	isActive bool,
	createdAt int64,
	creator string,
) (Tunnel, error) {
	msg, ok := route.(proto.Message)
	if !ok {
		return Tunnel{}, fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return Tunnel{}, err
	}

	return Tunnel{
		ID:               id,
		Sequence:         sequence,
		Route:            any,
		Encoder:          encoder,
		FeePayer:         feePayer,
		SignalDeviations: signalDeviations,
		Interval:         interval,
		TotalDeposit:     totalDeposit,
		IsActive:         isActive,
		CreatedAt:        createdAt,
		Creator:          creator,
	}, nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (t Tunnel) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var route RouteI
	return unpacker.UnpackAny(t.Route, &route)
}

// SetRoute sets the route of the tunnel.
func (t *Tunnel) SetRoute(route RouteI) error {
	msg, ok := route.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return err
	}
	t.Route = any

	return nil
}

// GetSignalDeviationMap returns the signal deviation map of the tunnel.
func (t Tunnel) GetSignalDeviationMap() map[string]SignalDeviation {
	signalDeviationMap := make(map[string]SignalDeviation, len(t.SignalDeviations))
	for _, sd := range t.SignalDeviations {
		signalDeviationMap[sd.SignalID] = sd
	}
	return signalDeviationMap
}

// NewIBCRoute creates a new IBCRoute instance.
func NewIBCRoute(channelID string) *IBCRoute {
	return &IBCRoute{
		ChannelID: channelID,
	}
}

// NewIBCPacketContent creates a new IBCPacketContent instance.
func NewIBCPacketContent(channelID string) *IBCPacketContent {
	return &IBCPacketContent{
		ChannelID: channelID,
	}
}
