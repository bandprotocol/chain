package types

import (
	"fmt"

	proto "github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
)

var _ types.UnpackInterfacesMessage = Packet{}

func NewPacket(
	tunnelID uint64,
	sequence uint64,
	prices []feedstypes.Price,
	createdAt int64,
) Packet {
	return Packet{
		TunnelID:  tunnelID,
		Sequence:  sequence,
		Prices:    prices,
		Receipt:   nil,
		CreatedAt: createdAt,
	}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (p Packet) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var receipt PacketReceiptI
	return unpacker.UnpackAny(p.Receipt, &receipt)
}

// SetReceipt sets the packet's receipt.
func (p *Packet) SetReceipt(receipt PacketReceiptI) error {
	msg, ok := receipt.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msg)
	}

	any, err := types.NewAnyWithValue(receipt)
	if err != nil {
		return err
	}
	p.Receipt = any

	return nil
}

// GetReceiptValue returns the packet's receipt.
func (p Packet) GetReceiptValue() (PacketReceiptI, error) {
	r, ok := p.Receipt.GetCachedValue().(PacketReceiptI)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf(
			"expected %T, got %T",
			(PacketReceiptI)(nil),
			p.Receipt.GetCachedValue(),
		)
	}

	return r, nil
}
