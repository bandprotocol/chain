package types

import (
	"bytes"

	"github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
)

var (
	_ Originator = &DirectOriginator{}
	_ Originator = &TunnelOriginator{}
)

const (
	DirectOriginatorPrefix = "\xb3\x9f\xa5\xd2" // tss.Hash([]byte("DirectOriginator"))[:4]
	TunnelOriginatorPrefix = "\x72\xeb\xe8\x3d" // tss.Hash([]byte("TunnelOriginator"))[:4]
)

// Originator is the interface for identifying the metadata of the message. The hashed of the
// encoded originator will be included as a part of a signed message.
type Originator interface {
	proto.Message

	Encode() ([]byte, error)
	Validate(p Params) error
	Type() string
}

// ====================================
// DirectOriginator
// ====================================

// NewDirectOriginator creates a new direct originator.
func NewDirectOriginator(sourceChainID, requester, memo string) DirectOriginator {
	return DirectOriginator{
		SourceChainID: sourceChainID,
		Requester:     requester,
		Memo:          memo,
	}
}

// Validate checks the validity of the originator.
func (o DirectOriginator) Validate(p Params) error {
	if o.SourceChainID == "" {
		return ErrInvalidOriginator.Wrap("source chain ID cannot be empty")
	}

	if o.Requester == "" {
		return ErrInvalidOriginator.Wrap("requester cannot be empty")
	}

	if uint64(len(o.Memo)) > p.MaxMemoLength {
		return ErrInvalidOriginator.Wrapf("memo length exceeds maximum length of %d", p.MaxMemoLength)
	}

	return nil
}

// Encode encodes the originator into a byte array.
func (o DirectOriginator) Encode() ([]byte, error) {
	bz := bytes.Join([][]byte{
		[]byte(DirectOriginatorPrefix),
		tss.Hash([]byte(o.SourceChainID)),
		tss.Hash([]byte(o.Requester)),
		tss.Hash([]byte(o.Memo)),
	}, []byte(""))

	return bz, nil
}

// Type returns the type of the originator.
func (o DirectOriginator) Type() string {
	return "DirectOriginator"
}

// ====================================
// TunnelOriginator
// ====================================

// NewTunnelOriginator creates a new tunnel originator.
func NewTunnelOriginator(
	sourceChainID string,
	tunnelID uint64,
	destinationChainID string,
	destinationContractAddress string,
) TunnelOriginator {
	return TunnelOriginator{
		SourceChainID:              sourceChainID,
		TunnelID:                   tunnelID,
		DestinationChainID:         destinationChainID,
		DestinationContractAddress: destinationContractAddress,
	}
}

// Validate checks the validity of the originator.
func (o TunnelOriginator) Validate(p Params) error {
	if o.SourceChainID == "" {
		return ErrInvalidOriginator.Wrap("source chain ID cannot be empty")
	}

	if o.TunnelID == 0 {
		return ErrInvalidOriginator.Wrap("tunnel ID cannot be zero")
	}

	if o.DestinationContractAddress == "" {
		return ErrInvalidOriginator.Wrap("destination contract address cannot be empty")
	}

	if o.DestinationChainID == "" {
		return ErrInvalidOriginator.Wrap("destination chain ID cannot be empty")
	}

	return nil
}

// Encode encodes the originator into a byte array.
func (o TunnelOriginator) Encode() ([]byte, error) {
	bz := bytes.Join([][]byte{
		[]byte(TunnelOriginatorPrefix),
		tss.Hash([]byte(o.SourceChainID)),
		sdk.Uint64ToBigEndian(o.TunnelID),
		tss.Hash([]byte(o.DestinationChainID)),
		tss.Hash([]byte(o.DestinationContractAddress)),
	}, []byte(""))

	return bz, nil
}

// Type returns the type of the originator.
func (o TunnelOriginator) Type() string {
	return "TunnelOriginator"
}
