package feeds

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/feeds/keeper"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

var (
	FeedTypeFixedPointABIPrefix = tss.Hash([]byte("fixedPointABI"))[:4]
	FeedTypeTickABIPrefix       = tss.Hash([]byte("tickABI"))[:4]
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
			fp, err := k.GetFeedsPriceData(ctx, c.SignalIDs, c.FeedType)
			if err != nil {
				return nil, err
			}

			// Encode feeds price data
			bz, err := fp.ABIEncode()
			if err != nil {
				return nil, err
			}

			// Append the prefix based on the feed type
			switch c.FeedType {
			case types.FEED_TYPE_FIXED_POINT_ABI:
				return append(FeedTypeFixedPointABIPrefix, bz...), nil
			case types.FEED_TYPE_TICK_ABI:
				return append(FeedTypeTickABIPrefix, bz...), nil
			default:
				return nil, sdkerrors.ErrUnknownRequest.Wrapf(
					"unrecognized feed type: %d",
					c.FeedType,
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
