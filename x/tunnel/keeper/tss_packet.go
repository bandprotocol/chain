package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// SendTSSPacket sends TSS packet
func (k Keeper) SendTSSPacket(
	ctx sdk.Context,
	route *types.TSSRoute,
	packet types.Packet,
) (types.PacketContentI, error) {
	// Get SignalIDs from packet
	signalIDs := make([]string, len(packet.SignalPrices))
	for _, signalPrice := range packet.SignalPrices {
		signalIDs = append(signalIDs, signalPrice.SignalID)
	}

	tunnel := k.MustGetTunnel(ctx, packet.TunnelID)
	content := feedstypes.NewFeedSignatureOrder(signalIDs, tunnel.FeedType)
	feePerSigner := k.bandtssKeeper.GetParams(ctx).Fee

	// Sign TSS packet
	signingID, err := k.bandtssKeeper.CreateTunnelSigningRequest(
		ctx,
		packet.TunnelID,
		route.DestinationContractAddress,
		route.DestinationChainID,
		content,
		sdk.MustAccAddressFromBech32(tunnel.FeePayer),
		feePerSigner,
	)
	if err != nil {
		return nil, err
	}

	// Set the packet content
	packetContent := types.TSSPacketContent{
		SigningID:                  signingID,
		DestinationChainID:         route.DestinationChainID,
		DestinationContractAddress: route.DestinationContractAddress,
	}

	return &packetContent, nil
}
