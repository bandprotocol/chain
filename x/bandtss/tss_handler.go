package bandtss

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

const GroupTransitionMsgPrefix = "\x61\xb9\xb7\x41" // tss.Hash([]byte("Transition"))[:4]

// NewSignatureOrderHandler implements the Handler interface for tss module-based
// request signatures (ie. TextSignatureOrder)
func NewSignatureOrderHandler() tsstypes.Handler {
	return func(ctx sdk.Context, content tsstypes.Content) ([]byte, error) {
		switch c := content.(type) {
		case *types.GroupTransitionSignatureOrder:
			return bytes.Join(
				[][]byte{
					[]byte(GroupTransitionMsgPrefix),
					c.PubKey,
					sdk.Uint64ToBigEndian(uint64(c.TransitionTime.Unix())),
				},
				[]byte(""),
			), nil

		default:
			return nil, sdkerrors.ErrUnknownRequest.Wrapf(
				"unrecognized tss request signature message type: %s",
				c.OrderType(),
			)
		}
	}
}
