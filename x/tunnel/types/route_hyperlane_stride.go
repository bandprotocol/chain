package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// RouterRoute defines the Router route for the tunnel module
var _ RouteI = &HyperlaneStrideRoute{}

// NewRouterRoute creates a new RouterRoute instance.
func NewHyperlaneStrideRoute(
	dispatchDestDomain uint64,
	dispatchRecipientAddr string,
	fund sdk.Coin,
) *HyperlaneStrideRoute {
	return &HyperlaneStrideRoute{
		DispatchDestDomain:    dispatchDestDomain,
		DispatchRecipientAddr: dispatchRecipientAddr,
		Fund:                  fund,
	}
}

// ValidateBasic validates the HyperlaneStrideRoute
func (r *HyperlaneStrideRoute) ValidateBasic() error {
	// TODO: validate coin
	return nil
}

// NewHyperlaneStridePacketReceipt creates a new HyperlaneStridePacketReceipt instance.
func NewHyperlaneStridePacketReceipt(sequence uint64) *HyperlaneStridePacketReceipt {
	return &HyperlaneStridePacketReceipt{
		Sequence: sequence,
	}
}
