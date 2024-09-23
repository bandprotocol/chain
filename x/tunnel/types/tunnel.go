package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
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

// ValidateBasic performs basic validation of the LatestSignalPrices.
func (latestSignalPrices LatestSignalPrices) ValidateBasic() error {
	if latestSignalPrices.TunnelID == 0 {
		return fmt.Errorf("tunnel ID cannot be 0")
	}
	if len(latestSignalPrices.SignalPrices) == 0 {
		return fmt.Errorf("signal prices cannot be empty")
	}
	if latestSignalPrices.Timestamp < 0 {
		return fmt.Errorf("timestamp cannot be negative")
	}
	return nil
}
