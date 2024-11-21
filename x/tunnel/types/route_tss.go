package types

import bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"

var _ RouteI = &TSSRoute{}

// NewTSSRoute return a new TSSRoute instance.
func NewTSSRoute(
	destinationChainID, destinationContractAddress string,
) TSSRoute {
	return TSSRoute{
		DestinationChainID:         destinationChainID,
		DestinationContractAddress: destinationContractAddress,
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

	return nil
}

func NewTSSPacketContent(
	signingID bandtsstypes.SigningID,
	destinationChainID string,
	destinationContractAddress string,
) TSSPacketContent {
	return TSSPacketContent{
		SigningID:                  signingID,
		DestinationChainID:         destinationChainID,
		DestinationContractAddress: destinationContractAddress,
	}
}
