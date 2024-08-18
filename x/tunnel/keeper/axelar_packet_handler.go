package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// AxelarPacketHandler handles incoming Axelar packets
func (k Keeper) AxelarPacketHandler(ctx sdk.Context, route *types.AxelarRoute, packet types.Packet) {}
