package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	proto "github.com/cosmos/gogoproto/proto"

	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
)

var _ types.UnpackInterfacesMessage = Tunnel{}

// NewTunnel creates a new Tunnel instance.
func NewTunnel(
	id uint64,
	nonceCount uint64,
	route *codectypes.Any,
	feedType feedstypes.FeedType,
	feePayer string,
	signalPriceInfos []SignalPriceInfo,
	interval uint64,
	lastIntervalTimestamp int64,
	isActive bool,
	createdAt int64,
	creator string,
) Tunnel {
	return Tunnel{
		ID:                    id,
		NonceCount:            nonceCount,
		Route:                 route,
		FeedType:              feedType,
		FeePayer:              feePayer,
		SignalPriceInfos:      signalPriceInfos,
		Interval:              interval,
		LastIntervalTimestamp: lastIntervalTimestamp,
		IsActive:              isActive,
		CreatedAt:             createdAt,
		Creator:               creator,
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
