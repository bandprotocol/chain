package types

import (
	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// signature order types
const (
	SignatureOrderTypeTunnel = "Tunnel"
)

func init() {
	tsstypes.RegisterSignatureOrderTypeCodec(
		&TunnelSignatureOrder{},
		"tunnel/TunnelSignatureOrder",
	)
}

// Implements Content Interface
var _ tsstypes.Content = &TunnelSignatureOrder{}

// NewTunnelSignatureOrder returns a new TunnelSignatureOrder object
func NewTunnelSignatureOrder(packet Packet, feedType feedstypes.FeedType) *TunnelSignatureOrder {
	return &TunnelSignatureOrder{packet, feedType}
}

// OrderRoute returns the order router key
func (f *TunnelSignatureOrder) OrderRoute() string { return RouterKey }

// OrderType returns type of signature order that should be "Tunnel"
func (f *TunnelSignatureOrder) OrderType() string {
	return SignatureOrderTypeTunnel
}

// IsInternal returns true for TunnelSignatureOrder (internal module-based request signature).
func (f *TunnelSignatureOrder) IsInternal() bool { return true }

// ValidateBasic validates the request's title and description of the request signature
func (f *TunnelSignatureOrder) ValidateBasic() error { return nil }
