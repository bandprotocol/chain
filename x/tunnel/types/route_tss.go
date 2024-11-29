package types

import (
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
)

var _ RouteI = &TSSRoute{}

// NewTSSRoute return a new TSSRoute instance.
func NewTSSRoute(
	destinationChainID string,
	destinationContractAddress string,
	encoderType feedstypes.Encoder,
) TSSRoute {
	return TSSRoute{
		DestinationChainID:         destinationChainID,
		DestinationContractAddress: destinationContractAddress,
		Encoder:                    encoderType,
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

	if err := feedstypes.ValidateEncoder(r.Encoder); err != nil {
		return err
	}

	return nil
}

// NewTSSPacketReceipt creates a new TSSPacketReceipt instance.
func NewTSSPacketReceipt(signingID bandtsstypes.SigningID) *TSSPacketReceipt {
	return &TSSPacketReceipt{
		SigningID: signingID,
	}
}
