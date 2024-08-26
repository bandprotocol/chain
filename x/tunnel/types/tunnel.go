package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	proto "github.com/cosmos/gogoproto/proto"
)

var _ types.UnpackInterfacesMessage = Tunnel{}

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

// NewSignalPriceInfo creates a new SignalPriceInfo instance.
func NewSignalPriceInfo(
	signalID string,
	softDeviationBPS uint64,
	hardDeviationBPS uint64,
	price uint64,
	timestamp int64,
) SignalPriceInfo {
	return SignalPriceInfo{
		SignalID:         signalID,
		SoftDeviationBPS: softDeviationBPS,
		HardDeviationBPS: hardDeviationBPS,
		Price:            price,
		Timestamp:        timestamp,
	}
}
