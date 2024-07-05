package keeper

import (
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

// CalculateInterval calculates feed interval from power
func CalculateInterval(power int64, param types.Params) (interval int64) {
	if power < param.PowerStepThreshold {
		return 0
	}

	// divide power by power threshold to create steps
	powerFactor := power / param.PowerStepThreshold

	interval = max(param.MaxInterval/powerFactor, param.MinInterval)

	return
}

// CalculateDeviation calculates feed deviation from power
func CalculateDeviation(power int64, param types.Params) (deviation int64) {
	if power < param.PowerStepThreshold {
		return 0
	}

	// divide power by power threshold to create steps
	powerFactor := power / param.PowerStepThreshold

	deviation = max(param.MaxDeviationBasisPoint/powerFactor, param.MinDeviationBasisPoint)

	return
}

// sumPower sums power from a list of signals
func sumPower(signals []types.Signal) (sum int64) {
	for _, signal := range signals {
		sum += signal.Power
	}
	return
}
