package types

import "math"

// NewSignal creates a new signal
func NewSignal(id string, power int64) Signal {
	return Signal{
		ID:    id,
		Power: power,
	}
}

func (s *Signal) Validate() error {
	// Check if the signal ID is empty
	if s.ID == "" {
		return ErrInvalidSignal.Wrap(
			"signal id cannot be empty",
		)
	}

	// Check if the signal power is positive
	if s.Power <= 0 {
		return ErrInvalidSignal.Wrap(
			"signal power must be positive",
		)
	}

	// Check if the signal ID length exceeds the maximum allowed characters
	signalIDLength := len(s.ID)
	if uint64(signalIDLength) > MaxSignalIDCharacters {
		return ErrSignalIDTooLarge.Wrapf(
			"maximum number of characters is %d but received %d characters",
			MaxSignalIDCharacters, signalIDLength,
		)
	}

	return nil
}

// SumPower sums power from a list of signals.
// It returns ErrInvalidSignal if the total power would overflow int64.
func SumPower(signals []Signal) (int64, error) {
	var sum int64
	for _, signal := range signals {
		if signal.Power > 0 && sum > math.MaxInt64-signal.Power {
			return 0, ErrInvalidSignal.Wrap("total signal power overflows int64")
		}
		sum += signal.Power
	}
	return sum, nil
}
