package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// GroupTransitionPath is the reserved path for transition group msg
const GroupTransitionPath = "transition"

// GroupTransitionMsgPrefix is the prefix for transition group msg.
var GroupTransitionMsgPrefix = tss.Hash([]byte(GroupTransitionPath))[:4]

// Implements SignatureRequest Interface
var _ tsstypes.Content = &GroupTransitionSignatureOrder{}

func NewGroupTransitionSignatureOrder(pubKey []byte) *GroupTransitionSignatureOrder {
	return &GroupTransitionSignatureOrder{PubKey: pubKey}
}

// OrderRoute returns the order router key
func (rs *GroupTransitionSignatureOrder) OrderRoute() string { return RouterKey }

// OrderType of GroupTransitionSignatureOrder is "transition"
func (rs *GroupTransitionSignatureOrder) OrderType() string {
	return GroupTransitionPath
}

// ValidateBasic performs no-op for this type
func (rs *GroupTransitionSignatureOrder) ValidateBasic() error { return nil }

// NewSignatureOrderHandler implements the Handler interface for tss module-based
// request signatures (ie. TextSignatureOrder)
func NewSignatureOrderHandler() tsstypes.Handler {
	return func(ctx sdk.Context, content tsstypes.Content) ([]byte, error) {
		switch c := content.(type) {
		case *GroupTransitionSignatureOrder:
			return append(GroupTransitionMsgPrefix, c.PubKey...), nil

		default:
			return nil, sdkerrors.ErrUnknownRequest.Wrapf(
				"unrecognized tss request signature message type: %s",
				c.OrderType(),
			)
		}
	}
}
