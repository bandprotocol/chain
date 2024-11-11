package keeper

import (
	"math"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// SendTSSPacket sends TSS packet
func (k Keeper) SendTSSPacket(
	ctx sdk.Context,
	route *types.TSSRoute,
	packet types.Packet,
) (packetContent types.PacketContentI, fee sdk.Coins, err error) {
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
		feeLimits = append(feeLimits, sdk.NewCoin(coin.Denom, sdkmath.NewInt(math.MaxInt)))
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
		return nil, nil, err
	}

	// Set the packet content
	packetContent = &types.TSSPacketContent{
		SigningID:                  signingID,
		DestinationChainID:         route.DestinationChainID,
		DestinationContractAddress: route.DestinationContractAddress,
	}

	// TODO: return the actual fee that using in the route if possible
	fee, err = route.Fee()
	if err != nil {
		return nil, nil, err
	}

	return packetContent, fee, nil
}