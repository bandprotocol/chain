package types

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
)

var _ types.UnpackInterfacesMessage = Packet{}

func NewPacket(
	tunnelID uint64,
	sequence uint64,
	prices []feedstypes.Price,
	baseFee sdk.Coins,
	routeFee sdk.Coins,
	createdAt int64,
) Packet {
	return Packet{
		TunnelID:    tunnelID,
		Sequence:    sequence,
		Prices:      prices,
		RouteResult: nil,
		BaseFee:     baseFee,
		RouteFee:    routeFee,
		CreatedAt:   createdAt,
	}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (p Packet) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var routeResult RouteResultI
	return unpacker.UnpackAny(p.RouteResult, &routeResult)
}

// SetRouteResultValue sets the route result of the packet.
func (p *Packet) SetRouteResultValue(routeResult RouteResultI) error {
	any, err := types.NewAnyWithValue(routeResult)
	if err != nil {
		return err
	}
	p.RouteResult = any

	return nil
}

// GetRouteResultValue returns the route result of the packet.
func (p Packet) GetRouteResultValue() (RouteResultI, error) {
	routeResult, ok := p.RouteResult.GetCachedValue().(RouteResultI)
	if !ok {
		return nil, ErrNoRouteResult.Wrapf("tunnelID: %d, sequence: %d", p.TunnelID, p.Sequence)
	}

	return routeResult, nil
}
