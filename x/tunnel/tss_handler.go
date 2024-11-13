package tunnel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

var (
	// EncoderFixedPointABIPrefix is the prefix for fixed point ABI encoder
	// The value is tss.Hash([]byte("FixedPointABI"))[:4]
	EncoderFixedPointABIPrefix = tss.Hash([]byte("FixedPointABI"))[:4]
	// EncoderTickABIPrefix is the prefix for tick ABI encoder
	// The value is tss.Hash([]byte("TickABI"))[:4]
	EncoderTickABIPrefix = tss.Hash([]byte("TickABI"))[:4]
)

// NewSignatureOrderHandler creates a tss handler to handle tunnel signature order
func NewSignatureOrderHandler(k keeper.Keeper) tsstypes.Handler {
	return func(ctx sdk.Context, content tsstypes.Content) ([]byte, error) {
		switch c := content.(type) {
		case *types.TunnelSignatureOrder:
			var prefix []byte
			switch c.Encoder {
			case feedstypes.ENCODER_FIXED_POINT_ABI:
				prefix = EncoderFixedPointABIPrefix
			case feedstypes.ENCODER_TICK_ABI:
				prefix = EncoderTickABIPrefix
			default:
				return nil, types.ErrInvalidEncoder.Wrapf("invalid encoder: %s", c.Encoder)
			}

			bz, err := c.Packet.EncodeTss(c.DestinationChainID, c.DestinationContractAddress, c.Encoder)
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
