package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// TSSPacketHandle handles incoming TSS packets
func (k Keeper) TSSPacketHandle(ctx sdk.Context, route *types.TSSRoute, packet types.Packet) error {
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
		panic(fmt.Sprintf("failed to set packet content: %s", err))
	}

	// Save the signed TSS packet
	k.SetPacket(ctx, packet)
	return nil
}
