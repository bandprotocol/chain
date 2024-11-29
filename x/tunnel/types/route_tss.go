package types

import (
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
)

var _ RouteI = &TSSRoute{}

// NewTSSRoute return a new TSSRoute instance.
func NewTSSRoute(
	destinationChainID string,
	destinationContractAddress string,
	encoder TSSRouteEncoder,
) TSSRoute {
	return TSSRoute{
		DestinationChainID:         destinationChainID,
		DestinationContractAddress: destinationContractAddress,
		Encoder:                    encoder,
	}
}

// ValidateBasic performs basic validation of the TSSRoute fields.
func (r *TSSRoute) ValidateBasic() error {
	if r.DestinationChainID == "" {
		return ErrInvalidRoute.Wrapf("destination chain ID cannot be empty")
	}

	if r.DestinationContractAddress == "" {
		return ErrInvalidRoute.Wrapf("destination contract address cannot be empty")
	}

	if _, ok := TSSRouteEncoder_name[int32(r.Encoder)]; !ok || r.Encoder == TSS_ROUTE_ENCODER_UNSPECIFIED {
		return ErrInvalidEncoder.Wrapf("invalid encoder: %s", r.Encoder)
	}

	return nil
}

// NewTSSPacketReceipt creates a new TSSPacketReceipt instance.
func NewTSSPacketReceipt(signingID bandtsstypes.SigningID) *TSSPacketReceipt {
	return &TSSPacketReceipt{
		SigningID: signingID,
	}
}
