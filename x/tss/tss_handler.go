package tss

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v3/x/tss/keeper"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

const TextMsgPrefix = "\xb1\xf7\x60\x16" // tss.Hash([]byte("Text"))[:4]

// NewSignatureOrderHandler implements the Handler interface for tss module-based
// request signatures (i.e. TextSignatureOrder)
func NewSignatureOrderHandler(k keeper.Keeper) types.Handler {
	return func(ctx sdk.Context, content types.Content) ([]byte, error) {
		switch c := content.(type) {
		case *types.TextSignatureOrder:
			maxMessageLength := k.GetParams(ctx).MaxMessageLength
			if uint64(len(c.Message)) > maxMessageLength {
				return nil, types.ErrInvalidMessage.Wrapf(
					"message length exceeds maximum length of %d", maxMessageLength,
				)
			}

			return append([]byte(TextMsgPrefix), c.Message...), nil
		default:
			return nil, sdkerrors.ErrUnknownRequest.Wrapf(
				"unrecognized tss request signature message type: %s",
				c.OrderType(),
			)
		}
	}
}
