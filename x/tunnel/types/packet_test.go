package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestPacket_SetPacketContent(t *testing.T) {
	packet := types.NewPacket(1, 1, nil, nil, nil, 0)
	packetContent := &types.TSSPacketContent{DestinationChainID: "chain-1", DestinationContractAddress: "contract-1"}

	err := packet.SetPacketContent(packetContent)
	require.NoError(t, err)
	require.NotNil(t, packet.PacketContent)
	require.Equal(t, packetContent, packet.PacketContent.GetCachedValue())
}

func TestPacket_GetContent(t *testing.T) {
	packet := types.NewPacket(1, 1, nil, nil, nil, 0)
	packetContent := &types.TSSPacketContent{DestinationChainID: "chain-1", DestinationContractAddress: "contract-1"}

	err := packet.SetPacketContent(packetContent)
	require.NoError(t, err)

	content, err := packet.GetContent()
	require.NoError(t, err)
	require.NotNil(t, content)
	require.Equal(t, packetContent, content)
}

func TestPacket_UnpackInterfaces(t *testing.T) {
	packet := types.NewPacket(1, 1, nil, nil, nil, 0)
	packetContent := &types.TSSPacketContent{DestinationChainID: "chain-1", DestinationContractAddress: "contract-1"}

	err := packet.SetPacketContent(packetContent)
	require.NoError(t, err)

	unpacker := codectypes.NewInterfaceRegistry()
	err = packet.UnpackInterfaces(unpacker)
	require.NoError(t, err)
}
