package types

import (
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

// signature order types
const SignatureOrderTypeTunnel = "tunnel"

// Implements Content Interface
var _ tsstypes.Content = &TunnelSignatureOrder{}

// NewTunnelSignatureOrder returns a new TunnelSignatureOrder object
func NewTunnelSignatureOrder(
	sequence uint64,
	prices []feedstypes.Price,
	createdAt int64,
	encoder feedstypes.Encoder,
) *TunnelSignatureOrder {
	return &TunnelSignatureOrder{
		Sequence:  sequence,
		Prices:    prices,
		CreatedAt: createdAt,
		Encoder:   encoder,
	}
}

// OrderRoute returns the order router key
func (ts *TunnelSignatureOrder) OrderRoute() string { return RouterKey }

// OrderType returns type of signature order that should be "tunnel"
func (ts *TunnelSignatureOrder) OrderType() string {
	return SignatureOrderTypeTunnel
}

// IsInternal returns true for TunnelSignatureOrder (internal module-based request signature).
func (ts *TunnelSignatureOrder) IsInternal() bool { return true }

// ValidateBasic validates the request's title and description of the request signature
func (ts *TunnelSignatureOrder) ValidateBasic() error { return nil }
