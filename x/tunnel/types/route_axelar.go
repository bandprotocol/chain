package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewAxelarRoute creates a new AxelarRoute instance.
func NewAxelarRoute(
	destinationChainID string,
	destinationContractAddress string,
	fee sdk.Coin,
) *AxelarRoute {
	return &AxelarRoute{
		DestinationChainID:         ChainName(destinationChainID),
		DestinationContractAddress: destinationContractAddress,
		Fee:                        fee,
	}
}

// ValidateBasic validates the AxelarRoute.
func (r *AxelarRoute) ValidateBasic() error {
	if !r.Fee.IsValid() || !r.Fee.IsPositive() {
		return fmt.Errorf("invalid fee: %s", r.Fee)
	}

	if err := r.DestinationChainID.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid destination chain ID")
	}

	if err := ValidateString(r.DestinationContractAddress); err != nil {
		return errorsmod.Wrap(err, "invalid destination address")
	}

	return nil
}

// NewAxelarPacketReceipt creates a new AxelarPacketReceipt instance.
func NewAxelarPacketReceipt(sequence uint64) *AxelarPacketReceipt {
	return &AxelarPacketReceipt{
		Sequence: sequence,
	}
}
