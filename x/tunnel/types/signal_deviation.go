package types

// NewSignalDeviation creates a new SignalDeviation instance.
func NewSignalDeviation(
	signalID string,
	softDeviationBPS uint64,
	hardDeviationBPS uint64,
) SignalDeviation {
	return SignalDeviation{
		SignalID:         signalID,
		SoftDeviationBPS: softDeviationBPS,
		HardDeviationBPS: hardDeviationBPS,
	}
}

// ValidateSignalDeviations validates the signal deviations with the given params.
func ValidateSignalDeviations(
	signalDeviations []SignalDeviation,
	maxSignals uint64,
	maxDeviationBPS uint64,
	minDeviationBPS uint64,
) error {
	// validate max signals
	if len(signalDeviations) > int(maxSignals) {
		return ErrMaxSignalsExceeded.Wrapf("max signals %d, got %d", maxSignals, len(signalDeviations))
	}

	// validate min and max deviation
	for _, signalDeviation := range signalDeviations {
		if signalDeviation.HardDeviationBPS < minDeviationBPS ||
			signalDeviation.SoftDeviationBPS < minDeviationBPS ||
			signalDeviation.HardDeviationBPS > maxDeviationBPS ||
			signalDeviation.SoftDeviationBPS > maxDeviationBPS {
			return ErrDeviationOutOfRange.Wrapf(
				"min %d, max %d, got %d, %d",
				minDeviationBPS,
				maxDeviationBPS,
				signalDeviation.SoftDeviationBPS,
				signalDeviation.HardDeviationBPS,
			)
		}
	}
	return nil
}
