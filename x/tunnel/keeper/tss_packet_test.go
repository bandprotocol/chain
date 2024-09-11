package keeper_test

import (
	"math"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	bandtesting "github.com/bandprotocol/chain/v2/testing"
	bandtsstypes "github.com/bandprotocol/chain/v2/x/bandtss/types"
	"github.com/bandprotocol/chain/v2/x/tunnel/testutil"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestSendTSSPacket(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// Create a sample TSSRoute
	route := types.TSSRoute{
		DestinationChainID:         "chain-1",
		DestinationContractAddress: "0x1234567890abcdef",
	}

	// Create a sample Packet
	packet := types.NewPacket(
		1,                     // tunnelID
		1,                     // nonce
		[]types.SignalPrice{}, // SignalPriceInfos
		nil,
		time.Now().Unix(),
	)

	s.MockBandtssKeeper.EXPECT().GetParams(gomock.Any()).Return(bandtsstypes.Params{
		Fee: sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(10))),
	})
	s.MockBandtssKeeper.EXPECT().CreateTunnelSigningRequest(
		gomock.Any(),
		uint64(1),
		"0x1234567890abcdef",
		"chain-1",
		gomock.Any(),
		bandtesting.Alice.Address,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(math.MaxInt))),
	).Return(bandtsstypes.SigningID(1), nil)

	k.SetTunnel(ctx, types.Tunnel{
		ID:       1,
		Encoder:  types.ENCODER_FIXED_POINT_ABI,
		FeePayer: bandtesting.Alice.Address.String(),
	})

	// Send the TSS packet
	content, err := k.SendTSSPacket(ctx, &route, packet)
	require.NoError(t, err)

	// Assert the packet content
	packetContent, ok := content.(*types.TSSPacketContent)
	require.True(t, ok)
	require.Equal(t, "chain-1", packetContent.DestinationChainID)
	require.Equal(t, "0x1234567890abcdef", packetContent.DestinationContractAddress)
	require.Equal(t, bandtsstypes.SigningID(1), packetContent.SigningID)
}
