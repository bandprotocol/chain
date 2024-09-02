package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// HandleAxelarPacket handles Axelar packet
func (k Keeper) HandleAxelarPacket(
	ctx sdk.Context,
	route *types.AxelarRoute,
	packet types.Packet,
) (types.PacketContentI, error) {
	return nil, nil
}
