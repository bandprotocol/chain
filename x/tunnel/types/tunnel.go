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
	signalInfos []SignalInfo,
	interval uint64,
	totalDeposit []sdk.Coin,
	isActive bool,
	createdAt int64,
	creator string,
) Tunnel {
	return Tunnel{
		ID:           id,
		NonceCount:   nonceCount,
		Route:        route,
		Encoder:      encoder,
		FeePayer:     feePayer,
		SignalInfos:  signalInfos,
		Interval:     interval,
		TotalDeposit: totalDeposit,
		IsActive:     isActive,
		CreatedAt:    createdAt,
		Creator:      creator,
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

// GetSignalInfoMap returns the signal info map by signal ID from the tunnel.
func (t Tunnel) GetSignalInfoMap() map[string]SignalInfo {
	signalInfoMap := make(map[string]SignalInfo, len(t.SignalInfos))
	for _, si := range t.SignalInfos {
		signalInfoMap[si.SignalID] = si
	}
	return signalInfoMap
}
