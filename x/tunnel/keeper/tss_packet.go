package keeper

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// SendTSSPacket sends TSS packet
func (k Keeper) SendTSSPacket(
	ctx sdk.Context,
	route *types.TSSRoute,
	packet types.Packet,
) (types.PacketContentI, error) {
	tunnel := k.MustGetTunnel(ctx, packet.TunnelID)
	content := types.NewTunnelSignatureOrder(
		packet,
		route.DestinationChainID,
		route.DestinationContractAddress,
		tunnel.Encoder,
	)

	// assign feeLimit to infinite
	feePerSigner := k.bandtssKeeper.GetParams(ctx).Fee
	feeLimits := sdk.NewCoins()
	for _, coin := range feePerSigner {
		feeLimits = append(feeLimits, sdk.NewCoin(coin.Denom, sdk.NewInt(math.MaxInt)))
	}

	// Sign TSS packet
	signingID, err := k.bandtssKeeper.CreateTunnelSigningRequest(
		ctx,
		packet.TunnelID,
		route.DestinationContractAddress,
		route.DestinationChainID,
		content,
		sdk.MustAccAddressFromBech32(tunnel.FeePayer),
		feeLimits,
	)
	if err != nil {
		return nil, err
	}

	// Set the packet content
	packetContent := types.TSSPacketContent{
		SigningID:                  signingID,
		DestinationChainID:         route.DestinationChainID,
		DestinationContractAddress: route.DestinationContractAddress,
	}

	return &packetContent, nil
}
