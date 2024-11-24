package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestGetSetPacketReceipt(t *testing.T) {
	packet := types.NewPacket(1, 1, nil, nil, nil, 0)
	receipt := &types.TSSPacketReceipt{SigningID: 1}

	err := packet.SetReceipt(receipt)
	require.NoError(t, err)

	receiptValue, err := packet.GetReceiptValue()
	require.NoError(t, err)
	require.Equal(t, receipt, receiptValue)
}

func TestPacketUnpackInterfaces(t *testing.T) {
	packet := types.NewPacket(1, 1, nil, nil, nil, 0)
	packetResult := &types.TSSPacketReceipt{SigningID: 1}

	err := packet.SetReceipt(packetResult)
	require.NoError(t, err)

	unpacker := codectypes.NewInterfaceRegistry()
	err = packet.UnpackInterfaces(unpacker)
	require.NoError(t, err)
}
