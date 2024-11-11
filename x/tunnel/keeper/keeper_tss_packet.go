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
	// TODO: Implement TSS packet handler logic

	// Sign TSS packet

	// Set the packet content
	packetContent = &types.TSSPacketContent{
		SigningID:                  1,
		DestinationChainID:         route.DestinationChainID,
		DestinationContractAddress: route.DestinationContractAddress,
	}

	// TODO: return the actual fee that using in the route if possible
	fee, err = route.Fee()
	if err != nil {
		return nil, nil, err
	}

	return
}
