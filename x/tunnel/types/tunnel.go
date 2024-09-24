package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

var _ types.UnpackInterfacesMessage = Tunnel{}

// NewTunnel creates a new Tunnel instance.
func NewTunnel(
	id uint64,
	nonceCount uint64,
	route *types.Any,
	encoder Encoder,
	feePayer string,
	signalDeviations []SignalDeviation,
	interval uint64,
	totalDeposit []sdk.Coin,
	isActive bool,
	createdAt int64,
	creator string,
) Tunnel {
	return Tunnel{
		ID:               id,
		NonceCount:       nonceCount,
		Route:            route,
		Encoder:          encoder,
		FeePayer:         feePayer,
		SignalDeviations: signalDeviations,
		Interval:         interval,
		TotalDeposit:     totalDeposit,
		IsActive:         isActive,
		CreatedAt:        createdAt,
		Creator:          creator,
	}
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
