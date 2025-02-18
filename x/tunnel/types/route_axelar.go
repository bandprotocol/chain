package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewAxelarRoute creates a new AxelarRoute instance.
func NewAxelarRoute(
	destinationChainID string,
	destinationContractAddress string,
	fee sdk.Coin,
) *AxelarRoute {
	return &AxelarRoute{
		DestinationChainID:         destinationChainID,
		DestinationContractAddress: destinationContractAddress,
		Fee:                        fee,
	}
}

// ValidateBasic validates the AxelarRoute.
func (r *AxelarRoute) ValidateBasic() error {
	// Validate fee coin
	if !r.Fee.IsValid() && !r.Fee.IsPositive() {
		return fmt.Errorf("invalid fee: %s", r.Fee)
	}

	return nil
}

// NewAxelarPacketReceipt creates a new AxelarPacketReceipt instance.
func NewAxelarPacketReceipt(sequence uint64) *AxelarPacketReceipt {
	return &AxelarPacketReceipt{
		Sequence: sequence,
	}
}
