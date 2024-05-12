package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// ReplaceGroupPath is the reserved path for replace group msg
const ReplaceGroupPath = "replace"

// ReplaceGroupMsgPrefix is the prefix for replace group msg.
var ReplaceGroupMsgPrefix = tss.Hash([]byte(ReplaceGroupPath))[:4]

// Implements SignatureRequest Interface
var _ tsstypes.Content = &ReplaceGroupSignatureOrder{}

func NewReplaceGroupSignatureOrder(pubKey []byte) *ReplaceGroupSignatureOrder {
	return &ReplaceGroupSignatureOrder{PubKey: pubKey}
}

// OrderRoute returns the order router key
func (rs *ReplaceGroupSignatureOrder) OrderRoute() string { return RouterKey }

// OrderType of ReplaceGroupSignatureOrder is "replace"
func (rs *ReplaceGroupSignatureOrder) OrderType() string {
	return ReplaceGroupPath
}

// ValidateBasic performs no-op for this type
func (rs *ReplaceGroupSignatureOrder) ValidateBasic() error { return nil }

// NewSignatureOrderHandler implements the Handler interface for tss module-based
// request signatures (ie. TextSignatureOrder)
func NewSignatureOrderHandler() tsstypes.Handler {
	return func(ctx sdk.Context, content tsstypes.Content) ([]byte, error) {
		switch c := content.(type) {
		case *ReplaceGroupSignatureOrder:
			return append(ReplaceGroupMsgPrefix, c.PubKey...), nil

		default:
			return nil, sdkerrors.ErrUnknownRequest.Wrapf(
				"unrecognized tss request signature message type: %s",
				c.OrderType(),
			)
		}
	}
}
