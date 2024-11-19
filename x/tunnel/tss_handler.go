package tunnel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v3/x/feeds"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// NewSignatureOrderHandler creates a tss handler to handle tunnel signature order
func NewSignatureOrderHandler(k keeper.Keeper) tsstypes.Handler {
	return func(ctx sdk.Context, content tsstypes.Content) ([]byte, error) {
		switch c := content.(type) {
		case *types.TunnelSignatureOrder:
			tunnel, err := k.GetTunnel(ctx, c.TunnelID)
			if err != nil {
				return nil, err
			}

			route, ok := tunnel.Route.GetCachedValue().(*types.TSSRoute)
			if !ok {
				return nil, types.ErrInvalidRoute.Wrap("invalid route type; expect TSSRoute type")
			}

			packet, err := k.GetPacket(ctx, c.TunnelID, c.Sequence)
			if err != nil {
				return nil, err
			}

			var prefix []byte
			switch tunnel.Encoder {
			case feedstypes.ENCODER_FIXED_POINT_ABI:
				prefix = feeds.EncoderFixedPointABIPrefix
			case feedstypes.ENCODER_TICK_ABI:
				prefix = feeds.EncoderTickABIPrefix
			default:
				return nil, types.ErrInvalidEncoder.Wrapf("invalid encoder: %s", tunnel.Encoder)
			}

			bz, err := packet.EncodeTss(route.DestinationChainID, route.DestinationContractAddress, tunnel.Encoder)
			if err != nil {
				return nil, err
			}

			return append(prefix, bz...), nil
		default:
			return nil, sdkerrors.ErrUnknownRequest.Wrapf(
				"unrecognized tss request signature type: %s",
				c.OrderType(),
			)
		}
	}
}
