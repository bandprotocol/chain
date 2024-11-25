package types

import (
	"time"

	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

// GroupTransitionPath is the reserved path for transition group msg
const GroupTransitionPath = "transition"

// Implements SignatureRequest Interface
var _ tsstypes.Content = &GroupTransitionSignatureOrder{}

func NewGroupTransitionSignatureOrder(
	pubKey []byte,
	transitionTime time.Time,
) *GroupTransitionSignatureOrder {
	return &GroupTransitionSignatureOrder{
		PubKey:         pubKey,
		TransitionTime: transitionTime,
	}
}

// OrderRoute returns the order router key
func (rs *GroupTransitionSignatureOrder) OrderRoute() string { return RouterKey }

// OrderType of GroupTransitionSignatureOrder is "transition"
func (rs *GroupTransitionSignatureOrder) OrderType() string {
	return GroupTransitionPath
}

// IsInternal returns true for GroupTransitionSignatureOrder (internal module-based request signature).
func (rs *GroupTransitionSignatureOrder) IsInternal() bool { return true }

// ValidateBasic performs no-op for this type
func (rs *GroupTransitionSignatureOrder) ValidateBasic() error { return nil }
