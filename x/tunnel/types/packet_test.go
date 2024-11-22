package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestGetSetPacketReceipt(t *testing.T) {
	packet := types.NewPacket(1, 1, nil, nil, nil, 0)
	routeResult := &types.TSSPacketReceipt{SigningID: 1}

	err := packet.SetReceiptValue(routeResult)
	require.NoError(t, err)
	require.NotNil(t, packet.Receipt)

	result, err := packet.GetReceiptValue()
	require.NoError(t, err)
	require.Equal(t, routeResult, result)
}

func TestPacketUnpackInterfaces(t *testing.T) {
	packet := types.NewPacket(1, 1, nil, nil, nil, 0)
	packetResult := &types.TSSPacketReceipt{SigningID: 1}

	err := packet.SetReceiptValue(packetResult)
	require.NoError(t, err)

	unpacker := codectypes.NewInterfaceRegistry()
	err = packet.UnpackInterfaces(unpacker)
	require.NoError(t, err)
}
