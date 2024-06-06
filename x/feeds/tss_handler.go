package feeds

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v2/x/feeds/keeper"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// NewSignatureOrderHandler creates a tss handler to handle feeds signature order
func NewSignatureOrderHandler(k keeper.Keeper) tsstypes.Handler {
	return func(ctx sdk.Context, content tsstypes.Content) ([]byte, error) {
		switch c := content.(type) {
		case *types.FeedsSignatureOrder:
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
			case types.FEED_TYPE_DEFAULT:
				return append(types.FeedTypeDefaultPrefix, bz...), nil
			case types.FEED_TYPE_TICK:
				return append(types.FeedTypeTickPrefix, bz...), nil
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
