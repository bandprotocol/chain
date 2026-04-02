package types

import sdk "github.com/cosmos/cosmos-sdk/types"

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
	ctx sdk.Context,
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
		if ctx.BlockHeader().ChainID == "bandchain" || (ctx.BlockHeader().ChainID == "band-v3-testnet-1" && ctx.BlockHeight() > 39985000) {
			if IsDeviationOutOfRange(signalDeviation, maxDeviationBPS, minDeviationBPS) {
				return ErrDeviationOutOfRange.Wrapf(
					"min %d, max %d, got %d, %d",
					minDeviationBPS,
					maxDeviationBPS,
					signalDeviation.SoftDeviationBPS,
					signalDeviation.HardDeviationBPS,
				)
			}
		} else {
			if IsDeviationOutOfRangeLegacy(signalDeviation, maxDeviationBPS, minDeviationBPS) {
				return ErrDeviationOutOfRange.Wrapf(
					"min %d, max %d, got %d, %d",
					minDeviationBPS,
					maxDeviationBPS,
					signalDeviation.SoftDeviationBPS,
					signalDeviation.HardDeviationBPS,
				)
			}
		}
	}
	return nil
}

func IsDeviationOutOfRangeLegacy(
	signalDeviation SignalDeviation,
	maxDeviationBPS uint64,
	minDeviationBPS uint64,
) bool {
	return signalDeviation.HardDeviationBPS < minDeviationBPS ||
		signalDeviation.SoftDeviationBPS < minDeviationBPS ||
		signalDeviation.HardDeviationBPS > maxDeviationBPS ||
		signalDeviation.SoftDeviationBPS > maxDeviationBPS
}

func IsDeviationOutOfRange(
	signalDeviation SignalDeviation,
	maxDeviationBPS uint64,
	minDeviationBPS uint64,
) bool {
	return signalDeviation.HardDeviationBPS < minDeviationBPS ||
		signalDeviation.HardDeviationBPS > maxDeviationBPS ||
		signalDeviation.SoftDeviationBPS > signalDeviation.HardDeviationBPS
}
