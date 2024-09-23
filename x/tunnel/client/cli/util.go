package cli

import (
	"encoding/json"
	"os"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// SignalDeviations represents the signal deviation in the file
type SignalDeviations struct {
	SignalDeviations []SignalDeviation `json:"signal_deviations"`
}

// SignalDeviation represents the signal information without soft deviation, which may be implemented in the future.
type SignalDeviation struct {
	SignalID     string `json:"signal_id"`
	DeviationBPS uint64 `json:"deviation_bps"`
}

// ToSignalDeviations converts signal information to types.SignalDeviation, excluding soft deviation.
// Note: Soft deviation may be utilized in the future for deviation adjustments.
func (sis SignalDeviations) ToSignalDeviations() []types.SignalDeviation {
	var signalDeviations []types.SignalDeviation
	for _, si := range sis.SignalDeviations {
		signalDeviation := types.SignalDeviation{
			SignalID:         si.SignalID,
			SoftDeviationBPS: si.DeviationBPS,
			HardDeviationBPS: si.DeviationBPS,
		}
		signalDeviations = append(signalDeviations, signalDeviation)
	}
	return signalDeviations
}

// parseSignalDeviations parses the signal infos from the given file
func parseSignalDeviations(signalDeviationsFile string) (SignalDeviations, error) {
	var signalDeviations SignalDeviations

	if signalDeviationsFile == "" {
		return SignalDeviations{}, nil
	}

	contents, err := os.ReadFile(signalDeviationsFile)
	if err != nil {
		return SignalDeviations{}, err
	}

	if err := json.Unmarshal(contents, &signalDeviations); err != nil {
		return SignalDeviations{}, err
	}

	return signalDeviations, nil
}
