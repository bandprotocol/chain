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
	feePayer sdk.AccAddress,
) (packetContent types.PacketContentI, fee sdk.Coins, err error) {
	content := types.NewTunnelSignatureOrder(packet.TunnelID, packet.Sequence)

	// Sign TSS packet
	signingID, err := k.bandtssKeeper.CreateTunnelSigningRequest(
		ctx,
		packet.TunnelID,
		route.DestinationContractAddress,
		route.DestinationChainID,
		content,
		feePayer,
		packet.RouteFee,
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

	return packetContent, packet.RouteFee, nil
}
