package types

import (
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

// signature order types
const (
	SignatureOrderTypeTunnel = "Tunnel"
)

// Implements Content Interface
var _ tsstypes.Content = &TunnelSignatureOrder{}

// NewTunnelSignatureOrder returns a new TunnelSignatureOrder object
func NewTunnelSignatureOrder(
	packet Packet,
	destinationChainID string,
	destinationContractAddress string,
	encoder Encoder,
) *TunnelSignatureOrder {
	return &TunnelSignatureOrder{packet, destinationChainID, destinationContractAddress, encoder}
}

// OrderRoute returns the order router key
func (ts *TunnelSignatureOrder) OrderRoute() string { return RouterKey }

// OrderType returns type of signature order that should be "Tunnel"
func (ts *TunnelSignatureOrder) OrderType() string {
	return SignatureOrderTypeTunnel
}

// IsInternal returns true for TunnelSignatureOrder (internal module-based request signature).
func (ts *TunnelSignatureOrder) IsInternal() bool { return true }

// ValidateBasic validates the request's title and description of the request signature
func (ts *TunnelSignatureOrder) ValidateBasic() error { return nil }
