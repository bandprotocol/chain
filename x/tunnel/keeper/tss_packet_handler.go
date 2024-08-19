package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// TSSPacketHandler handles incoming TSS packets
func (k Keeper) TSSPacketHandler(ctx sdk.Context, route *types.TSSRoute, packet types.Packet) {
	// TODO: Implement TSS packet handler logic
	// Sign TSS packet

	// Set the packet content
	packetContent := types.TSSPacketContent{
		SigningID:                  1,
		DestinationChainID:         route.DestinationChainID,
		DestinationContractAddress: route.DestinationContractAddress,
	}
	err := packet.SetPacketContent(&packetContent)
	if err != nil {
		panic(fmt.Errorf("failed to set packet content: %w", err))
	}

	// Save the signed TSS packet
	k.AddPacket(ctx, packet)
}
