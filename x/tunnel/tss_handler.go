package tunnel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
	"github.com/bandprotocol/chain/v2/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

var (
	EncodeTypeFixedPointABIPrefix = tss.Hash([]byte("FixedPointABI"))[:4]
	EncodeTypeTickABIPrefix       = tss.Hash([]byte("TickABI"))[:4]
)

// NewSignatureOrderHandler creates a tss handler to handle feeds signature order
func NewSignatureOrderHandler(k keeper.Keeper) tsstypes.Handler {
	return func(ctx sdk.Context, content tsstypes.Content) ([]byte, error) {
		switch c := content.(type) {
		case *types.TunnelSignatureOrder:
			switch c.Encoder {
			case types.ENCODER_FIXED_POINT_ABI:
				tssPacket, err := types.NewTssPacket(c.Packet, c.Encoder)
				if err != nil {
					return nil, err
				}

				bz, err := tssPacket.EncodeAbi()
				if err != nil {
					return nil, err
				}

				return append(EncodeTypeFixedPointABIPrefix, bz...), nil
			case types.ENCODER_TICK_ABI:
				tssPacket, err := types.NewTssPacket(c.Packet, c.Encoder)
				if err != nil {
					return nil, err
				}

				bz, err := tssPacket.EncodeAbi()
				if err != nil {
					return nil, err
				}

				return append(EncodeTypeTickABIPrefix, bz...), nil
			default:
				return nil, sdkerrors.ErrUnknownRequest.Wrapf(
					"unrecognized encoder mode: %s",
					c.Encoder.String(),
				)
			}
		default:
			return nil, sdkerrors.ErrUnknownRequest.Wrapf(
				"unrecognized tss request signature type: %s",
				c.OrderType(),
			)
		}
	}
}
