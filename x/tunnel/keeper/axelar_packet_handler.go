package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// SendAxelarPacket sends Axelar packet
func (k Keeper) SendAxelarPacket(
	ctx sdk.Context,
	route *types.AxelarRoute,
	packet types.Packet,
) (types.PacketContentI, error) {
	return nil, nil
}
