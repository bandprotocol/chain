package tunnel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
	"github.com/bandprotocol/chain/v2/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

var EncodeTypePacketABIPrefix = tss.Hash([]byte("PacketABI"))[:4]

// NewSignatureOrderHandler creates a tss handler to handle feeds signature order
func NewSignatureOrderHandler(k keeper.Keeper) tsstypes.Handler {
	return func(ctx sdk.Context, content tsstypes.Content) ([]byte, error) {
		switch c := content.(type) {
		case *types.TunnelSignatureOrder:
			tssPacket := types.NewTssPacket(c.Packet)
			bz, err := tssPacket.EncodeAbi()
			if err != nil {
				return nil, err
			}

			return append(EncodeTypePacketABIPrefix, bz...), nil
		default:
			return nil, sdkerrors.ErrUnknownRequest.Wrapf(
				"unrecognized tss request signature type: %s",
				c.OrderType(),
			)
		}
	}
}
