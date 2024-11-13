package types

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/bandprotocol/chain/v3/pkg/tss"
)

var (
	_ Originator = &DirectOriginator{}
	_ Originator = &TunnelOriginator{}

	// DirectOriginatorPrefix is the prefix for the originator from direct signing request.
	// The value is tss.Hash([]byte("directOriginatorPrefix"))[:4]
	DirectOriginatorPrefix = tss.Hash([]byte("directOriginatorPrefix"))[:4]
	// TunnelOriginatorPrefix is the prefix for the originator from tunnel module.
	// The value is tss.Hash([]byte("tunnelOriginatorPrefix"))[:4]
	TunnelOriginatorPrefix = tss.Hash([]byte("tunnelOriginatorPrefix"))[:4]
)

// Originator is the interface for identifying the metadata of the message. The hashed of the
// encoded originator will be included as a part of a signed message.
type Originator interface {
	proto.Message

	Encode() ([]byte, error)
	Validate(p Params) error
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
	if uint64(len(o.Memo)) > p.MaxMemoLength {
		return ErrInvalidMemo
	}

	return nil
}

// Encode encodes the originator into a byte array.
func (o DirectOriginator) Encode() ([]byte, error) {
	bz := bytes.Join([][]byte{
		DirectOriginatorPrefix,
		tss.Hash([]byte(o.SourceChainID)),
		tss.Hash([]byte(o.Requester)),
		tss.Hash([]byte(o.Memo)),
	}, []byte(""))

	return bz, nil
}

// ====================================
// TunnelOriginator
// ====================================

// NewTunnelOriginator creates a new tunnel originator.
func NewTunnelOriginator(
	sourceChainID string,
	tunnelID uint64,
	contractAddress, targetChainID string,
) TunnelOriginator {
	return TunnelOriginator{
		SourceChainID:   sourceChainID,
		TunnelID:        tunnelID,
		ContractAddress: contractAddress,
		TargetChainID:   targetChainID,
	}
}

// Validate checks the validity of the originator.
func (o TunnelOriginator) Validate(p Params) error {
	return nil
}

// Encode encodes the originator into a byte array.
func (o TunnelOriginator) Encode() ([]byte, error) {
	bz := bytes.Join([][]byte{
		TunnelOriginatorPrefix,
		tss.Hash([]byte(o.SourceChainID)),
		sdk.Uint64ToBigEndian(o.TunnelID),
		tss.Hash([]byte(o.ContractAddress)),
		tss.Hash([]byte(o.TargetChainID)),
	}, []byte(""))

	return bz, nil
}
