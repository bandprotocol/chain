package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestGetSetRouteResultValue(t *testing.T) {
	packet := types.NewPacket(1, 1, nil, nil, nil, 0)
	routeResult := &types.TSSRouteResult{SigningID: 1}

	err := packet.SetRouteResultValue(routeResult)
	require.NoError(t, err)
	require.NotNil(t, packet.RouteResult)

	result, err := packet.GetRouteResultValue()
	require.NoError(t, err)
	require.Equal(t, routeResult, result)
}

func TestRouteUnpackInterfaces(t *testing.T) {
	packet := types.NewPacket(1, 1, nil, nil, nil, 0)
	packetResult := &types.TSSRouteResult{SigningID: 1}

	err := packet.SetRouteResultValue(packetResult)
	require.NoError(t, err)

	unpacker := codectypes.NewInterfaceRegistry()
	err = packet.UnpackInterfaces(unpacker)
	require.NoError(t, err)
}
