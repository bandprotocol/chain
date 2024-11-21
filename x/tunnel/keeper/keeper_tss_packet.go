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
) (types.RouteResultI, error) {
	// TODO: Implement TSS packet handler logic

	// Sign TSS packet

	return &types.TSSRouteResult{
		SigningID: 1,
	}, nil
}
