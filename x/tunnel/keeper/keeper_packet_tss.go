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
	feePayer sdk.AccAddress,
) (receipt types.PacketReceiptI, err error) {
	content := types.NewTunnelSignatureOrder(packet.TunnelID, packet.Sequence)

	// try signing TSS packet, if success, write the context.
	signingID, err := k.bandtssKeeper.CreateTunnelSigningRequest(
		ctx,
		packet.TunnelID,
		route.DestinationChainID,
		route.DestinationContractAddress,
		content,
		feePayer,
		packet.RouteFee,
	)
	if err != nil {
		return nil, err
	}

	return types.NewTSSPacketReceipt(signingID), nil
}
