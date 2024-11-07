package types

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewVote creates a new Vote instance.
func NewVote(voter string, signals []Signal) Vote {
	return Vote{
		Voter:   voter,
		Signals: signals,
	}
}

// Validate validates the vote
func (v *Vote) Validate() error {
	if _, err := sdk.AccAddressFromBech32(v.Voter); err != nil {
		return errorsmod.Wrap(err, "invalid voter address")
	}

	// Map to track signal IDs for duplicate check
	signalIDSet := make(map[string]struct{})

	for _, signal := range v.Signals {
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
