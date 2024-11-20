package tunnel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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

			return types.EncodeTss(
				packet,
				route.DestinationChainID,
				route.DestinationContractAddress,
				tunnel.Encoder,
			)
		default:
			return nil, sdkerrors.ErrUnknownRequest.Wrapf(
				"unrecognized tss request signature type: %s",
				c.OrderType(),
			)
		}
	}
}
