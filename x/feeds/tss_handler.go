package feeds

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/feeds/keeper"
	"github.com/bandprotocol/chain/v3/x/feeds/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

var (
	// EncoderFixedPointABIPrefix is the constant prefix for feeds signature order on fixed point
	// ABI encoder message. The value is tss.Hash([]byte("fixedPointABI"))[:4]
	EncoderFixedPointABIPrefix = tss.Hash([]byte("fixedPointABI"))[:4]
	// EncoderTickABIPrefix is the constant prefix for feeds signature order on tick
	// ABI encoder message. The value is tss.Hash([]byte("tickABI"))[:4]
	EncoderTickABIPrefix = tss.Hash([]byte("tickABI"))[:4]
)

// NewSignatureOrderHandler creates a tss handler to handle feeds signature order
func NewSignatureOrderHandler(k keeper.Keeper) tsstypes.Handler {
	return func(ctx sdk.Context, content tsstypes.Content) ([]byte, error) {
		switch c := content.(type) {
		case *types.FeedsSignatureOrder:
			maxSignalIDs := k.GetParams(ctx).MaxSignalIDsPerSigning
			if uint64(len(c.SignalIDs)) > maxSignalIDs {
				return nil, types.ErrInvalidSignalIDs.Wrapf(
					"number of signal IDs exceeds maximum number of %d", maxSignalIDs,
				)
			}

			prices := k.GetPrices(ctx, c.SignalIDs)

			priceEncoders, err := types.ToPriceEncoders(prices, c.Encoder)
			if err != nil {
				return nil, err
			}

			// Append the prefix based on the encoder mode
			switch c.Encoder {
			case types.ENCODER_FIXED_POINT_ABI:
				bz, err := priceEncoders.EncodeABI(uint64(ctx.BlockTime().Unix()))
				if err != nil {
					return nil, err
				}
				return append(EncoderFixedPointABIPrefix, bz...), nil
			case types.ENCODER_TICK_ABI:
				bz, err := priceEncoders.EncodeABI(uint64(ctx.BlockTime().Unix()))
				if err != nil {
					return nil, err
				}
				return append(EncoderTickABIPrefix, bz...), nil
			default:
				return nil, sdkerrors.ErrUnknownRequest.Wrapf(
					"unrecognized encoder: %d",
					c.Encoder,
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
