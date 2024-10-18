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
) (types.PacketContentI, error) {
	// TODO: Implement TSS packet handler logic

	// Sign TSS packet

	// Set the packet content
	packetContent := types.TSSPacketContent{
		SigningID:                  1,
		DestinationChainID:         route.DestinationChainID,
		DestinationContractAddress: route.DestinationContractAddress,
	}

	return &packetContent, nil
}
