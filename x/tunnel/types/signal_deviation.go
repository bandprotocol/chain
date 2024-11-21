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
func ValidateSignalDeviations(signalDeviations []SignalDeviation, params Params) error {
	// validate max signals
	if len(signalDeviations) > int(params.MaxSignals) {
		return ErrMaxSignalsExceeded.Wrapf("max signals %d, got %d", params.MaxSignals, len(signalDeviations))
	}

	// validate min and max deviation
	for _, signalDeviation := range signalDeviations {
		if signalDeviation.HardDeviationBPS < params.MinDeviationBPS ||
			signalDeviation.SoftDeviationBPS < params.MinDeviationBPS ||
			signalDeviation.HardDeviationBPS > params.MaxDeviationBPS ||
			signalDeviation.SoftDeviationBPS > params.MaxDeviationBPS {
			return ErrDeviationOutOfRange.Wrapf(
				"min %d, max %d, got %d, %d",
				params.MinDeviationBPS,
				params.MaxDeviationBPS,
				signalDeviation.SoftDeviationBPS,
				signalDeviation.HardDeviationBPS,
			)
		}
	}
	return nil
}
