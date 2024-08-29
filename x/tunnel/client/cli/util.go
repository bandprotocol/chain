package cli

import (
	"encoding/json"
	"os"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// SignalInfos represents the signal infos in the file
type SignalInfos struct {
	SignalInfos []SignalInfo `json:"signal_infos"`
}

// SignalInfo represents the signal information without soft deviation, which may be implemented in the future.
type SignalInfo struct {
	SignalID     string `json:"signal_id"`
	DeviationBPS uint64 `json:"deviation_bps"`
}

// ToSignalInfos converts signal information to types.SignalInfo, excluding soft deviation.
// Note: Soft deviation may be utilized in the future for deviation adjustments.
func (sis SignalInfos) ToSignalInfos() []types.SignalInfo {
	var signalInfos []types.SignalInfo
	for _, si := range sis.SignalInfos {
		signalInfo := types.SignalInfo{
			SignalID:         si.SignalID,
			SoftDeviationBPS: 0,
			HardDeviationBPS: si.DeviationBPS,
		}
		signalInfos = append(signalInfos, signalInfo)
	}
	return signalInfos
}

// parseSignalInfos parses the signal infos from the given file
func parseSignalInfos(signalInfosFile string) (SignalInfos, error) {
	var signalInfos SignalInfos

	if signalInfosFile == "" {
		return SignalInfos{}, nil
	}

	contents, err := os.ReadFile(signalInfosFile)
	if err != nil {
		return SignalInfos{}, err
	}

	if err := json.Unmarshal(contents, &signalInfos); err != nil {
		return SignalInfos{}, err
	}

	return signalInfos, nil
}
