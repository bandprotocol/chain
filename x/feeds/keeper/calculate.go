package keeper

import "github.com/bandprotocol/chain/v2/x/feeds/types"

func calculateInterval(power int64, param types.Params) int64 {
	if power < param.PowerThreshold {
		return 0
	}

	// divide power by power threshold to create steps
	interval := param.MaxInterval / (power / param.PowerThreshold)
	if interval < param.MinInterval {
		return param.MinInterval
	}
	return interval
}

func sumPower(signals []types.Signal) (sum uint64) {
	for _, signal := range signals {
		sum += signal.Power
	}
	return
}
