package feeds

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v3/x/feeds/keeper"
	"github.com/bandprotocol/chain/v3/x/feeds/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

// NewSignatureOrderHandler creates a TSS handler to handle feeds signature order
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

			return types.EncodeTSS(prices, ctx.BlockTime().Unix(), c.Encoder)
		default:
			return nil, sdkerrors.ErrUnknownRequest.Wrapf(
				"unrecognized TSS request signature type: %s",
				c.OrderType(),
			)
		}
	}
}
