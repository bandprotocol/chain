package types

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

var (
	_ Originator = &DirectOriginator{}
	_ Originator = &TunnelOriginator{}

	directOriginatorPrefix = tss.Hash([]byte("directOriginatorPrefix"))[:4]
	tunnelOriginatorPrefix = tss.Hash([]byte("tunnelOriginatorPrefix"))[:4]
)

// Originator is the interface for identifying the metadata of the message. The hashed of the
// encoded originator will be included as a part of a signed message.
type Originator interface {
	Encode() ([]byte, error)
	Validate(p Params) error
}

func (o DirectOriginator) Validate(p Params) error {
	if uint64(len(o.Memo)) > p.MaxMemoLength {
		return ErrInvalidMemo
	}

	return nil
}

func (o DirectOriginator) Encode() ([]byte, error) {
	bz, err := marshal(&o)
	if err != nil {
		return nil, err
	}
	return append(directOriginatorPrefix, bz...), nil
}

func (o TunnelOriginator) Validate(p Params) error {
	return nil
}

func (o TunnelOriginator) Encode() ([]byte, error) {
	bz, err := marshal(&o)
	if err != nil {
		return nil, err
	}
	return append(tunnelOriginatorPrefix, bz...), nil
}

func marshal(pm codec.ProtoMarshaler) ([]byte, error) {
	// Size() check can catch the typed nil value.
	if pm == nil || pm.Size() == 0 {
		// return empty bytes instead of nil, because nil has special meaning in places like store.Set
		return []byte{}, nil
	}
	return pm.Marshal()
}
