package types

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
)

var _ types.UnpackInterfacesMessage = Tunnel{}

// NewTunnel creates a new Tunnel instance.
func NewTunnel(
	id uint64,
	sequence uint64,
	route RouteI,
	encoder feedstypes.Encoder,
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

// GetSignalIDs returns the signal IDs of the tunnel.
func (t Tunnel) GetSignalIDs() []string {
	signalIDs := make([]string, 0, len(t.SignalDeviations))
	for _, sd := range t.SignalDeviations {
		signalIDs = append(signalIDs, sd.SignalID)
	}
	return signalIDs
}

// Validate validates the total fees
func (tf TotalFees) Validate() error {
	if !tf.TotalPacketFee.IsValid() {
		return fmt.Errorf("invalid total packet fee: %s", tf.TotalPacketFee)
	}
	return nil
}

// ValidateInterval validates the interval.
func ValidateInterval(interval uint64, params Params) error {
	if interval < params.MinInterval || interval > params.MaxInterval {
		return ErrIntervalOutOfRange.Wrapf(
			"max %d, min %d, got %d",
			params.MaxInterval,
			params.MinInterval,
			interval,
		)
	}
	return nil
}
