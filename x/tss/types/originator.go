package types

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
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

// ====================================
// DirectOriginator
// ====================================

// Validate checks the validity of the originator.
func (o DirectOriginator) Validate(p Params) error {
	if uint64(len(o.Memo)) > p.MaxMemoLength {
		return ErrInvalidMemo
	}

	return nil
}

// Encode encodes the originator into a byte array.
func (o DirectOriginator) Encode() ([]byte, error) {
	bz := bytes.Join([][]byte{
		directOriginatorPrefix,
		sdk.Uint64ToBigEndian(uint64(len(o.Requester))),
		[]byte(o.Requester),
		sdk.Uint64ToBigEndian(uint64(len(o.Memo))),
		[]byte(o.Memo),
	}, []byte(""))

	return bz, nil
}

// ====================================
// TunnelOriginator
// ====================================

// Validate checks the validity of the originator.
func (o TunnelOriginator) Validate(p Params) error {
	return nil
}

// Encode encodes the originator into a byte array.
func (o TunnelOriginator) Encode() ([]byte, error) {
	bz := bytes.Join([][]byte{
		tunnelOriginatorPrefix,
		sdk.Uint64ToBigEndian(o.TunnelID),
		sdk.Uint64ToBigEndian(uint64(len(o.ContractAddress))),
		[]byte(o.ContractAddress),
		sdk.Uint64ToBigEndian(uint64(len(o.ChainID))),
		[]byte(o.ChainID),
	}, []byte(""))

	return bz, nil
}
