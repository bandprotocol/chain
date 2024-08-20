package types

// CalculateInterval calculates feed interval from power
func CalculateInterval(power int64, powerStep int64, minInterval int64, maxInterval int64) (interval int64) {
	if power < powerStep {
		return 0
	}

	// divide power by power threshold to create steps
	powerFactor := power / powerStep

	interval = max(maxInterval/powerFactor, minInterval)

	return
}

// CalculateDeviation calculates feed deviation from power
func CalculateDeviation(power int64, powerStep int64, minDeviationBP int64, maxDeviationBP int64) (deviation int64) {
	if power < powerStep {
		return 0
	}

	// divide power by power threshold to create steps
	powerFactor := power / powerStep

	deviation = max(maxDeviationBP/powerFactor, minDeviationBP)

	return
}
