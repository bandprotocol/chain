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
	EncoderFixedPointABIPrefix = tss.Hash([]byte("fixedPointABI"))[:4]
	EncoderTickABIPrefix       = tss.Hash([]byte("tickABI"))[:4]
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

			// Get feeds price data
			fp, err := k.GetFeedsPriceData(ctx, c.SignalIDs, c.Encoder)
			if err != nil {
				return nil, err
			}

			// Encode feeds price data
			bz, err := fp.ABIEncode()
			if err != nil {
				return nil, err
			}

			// Append the prefix based on the encoder mode
			switch c.Encoder {
			case types.ENCODER_FIXED_POINT_ABI:
				return append(EncoderFixedPointABIPrefix, bz...), nil
			case types.ENCODER_TICK_ABI:
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
