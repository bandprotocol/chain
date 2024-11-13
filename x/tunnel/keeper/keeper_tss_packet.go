package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// SendTSSPacket sends TSS packet
func (k Keeper) SendTSSPacket(
	ctx sdk.Context,
	route *types.TSSRoute,
	packet types.Packet,
) (packetContent types.PacketContentI, fee sdk.Coins, err error) {
	tunnel := k.MustGetTunnel(ctx, packet.TunnelID)
	content := types.NewTunnelSignatureOrder(
		packet,
		route.DestinationChainID,
		route.DestinationContractAddress,
		tunnel.Encoder,
	)

	fee, err = k.bandtssKeeper.GetSigningFee(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Sign TSS packet
	signingID, err := k.bandtssKeeper.CreateTunnelSigningRequest(
		ctx,
		packet.TunnelID,
		route.DestinationContractAddress,
		route.DestinationChainID,
		content,
		sdk.MustAccAddressFromBech32(tunnel.FeePayer),
		fee,
	)
	if err != nil {
		return nil, nil, err
	}

	// Set the packet content
	packetContent = &types.TSSPacketContent{
		SigningID:                  signingID,
		DestinationChainID:         route.DestinationChainID,
		DestinationContractAddress: route.DestinationContractAddress,
	}

	return packetContent, fee, nil
}
