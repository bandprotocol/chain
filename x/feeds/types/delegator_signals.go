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
	for _, signal := range ds.Signals {
		if err := validateString("signal ID", false, signal.ID); err != nil {
			return err
		}

		signalIDLength := len(signal.ID)
		if uint64(signalIDLength) > maxSignalIDCharacters {
			return ErrSignalIDTooLarge.Wrapf(
				"maximum number of characters is %d but received %d characters",
				maxSignalIDCharacters, signalIDLength,
			)
		}

		if err := validateInt64("signal power", true, signal.Power); err != nil {
			return err
		}
	}
	return nil
}
