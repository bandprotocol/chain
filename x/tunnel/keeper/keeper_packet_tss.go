package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// SendTSSPacket sends tss packet
func (k Keeper) SendTSSPacket(
	ctx sdk.Context,
	route *types.TSSRoute,
	packet types.Packet,
	feePayer sdk.AccAddress,
) (receipt types.PacketReceiptI, err error) {
	content := types.NewTunnelSignatureOrder(
		packet.Sequence,
		packet.Prices,
		packet.CreatedAt,
		route.Encoder,
	)

	tssFee, err := k.bandtssKeeper.GetSigningFee(ctx)
	if err != nil {
		return nil, err
	}

	// try signing tss packet, if success, write the context.
	signingID, err := k.bandtssKeeper.CreateTunnelSigningRequest(
		ctx,
		packet.TunnelID,
		route.DestinationChainID,
		route.DestinationContractAddress,
		content,
		feePayer,
		tssFee,
	)
	if err != nil {
		return nil, err
	}

	return types.NewTSSPacketReceipt(signingID), nil
}
