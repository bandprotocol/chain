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
		TunnelID:  tunnelID,
		Sequence:  sequence,
		Prices:    prices,
		Receipt:   nil,
		BaseFee:   baseFee,
		RouteFee:  routeFee,
		CreatedAt: createdAt,
	}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (p Packet) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var receipt PacketReceiptI
	return unpacker.UnpackAny(p.Receipt, &receipt)
}

// SetReceiptValue sets the packet's receipt.
func (p *Packet) SetReceiptValue(receipt PacketReceiptI) error {
	any, err := types.NewAnyWithValue(receipt)
	if err != nil {
		return err
	}
	p.Receipt = any

	return nil
}

// GetReceiptValue returns the packet's receipt.
func (p Packet) GetReceiptValue() (PacketReceiptI, error) {
	receipt, ok := p.Receipt.GetCachedValue().(PacketReceiptI)
	if !ok {
		return nil, ErrNoPacketReceipt.Wrapf("tunnelID: %d, sequence: %d", p.TunnelID, p.Sequence)
	}

	return receipt, nil
}
