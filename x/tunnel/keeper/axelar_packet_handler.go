package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// AxelarPacketHandle handles incoming Axelar packets
func (k Keeper) AxelarPacketHandle(ctx sdk.Context, route *types.AxelarRoute, packet types.Packet) error {
	return nil
}
