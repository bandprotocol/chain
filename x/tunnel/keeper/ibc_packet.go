package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// IBCPacketHandler func
func (k Keeper) IBCPacketHandler(ctx sdk.Context, packet types.IBCPacket) {}
