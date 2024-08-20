package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate validates the delegator signals
func (ds *DelegatorSignals) Validate(maxSignalIDCharacters uint64) error {
	if _, err := sdk.AccAddressFromBech32(ds.Delegator); err != nil {
		return errorsmod.Wrap(err, "invalid delegator address")
	}

	// Map to track signal IDs for duplicate check
	signalIDSet := make(map[string]struct{})

	for _, signal := range ds.Signals {
		// Validate signal ID
		if err := signal.Validate(); err != nil {
			return err
		}

		// Check for duplicate signal IDs
		if _, exists := signalIDSet[signal.ID]; exists {
			return ErrDuplicateSignalID.Wrapf(
				"duplicate signal ID found: %s", signal.ID,
			)
		}
		signalIDSet[signal.ID] = struct{}{}
	}
	return nil
}
