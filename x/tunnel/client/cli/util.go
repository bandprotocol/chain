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

// SignalInfo represents the signal info without soft deviation that will implement in the future
type SignalInfo struct {
	SignalID     string `json:"signal_id"`
	DeviationBPS uint64 `json:"deviation_bps"`
}

// ToSignalInfos converts the signal infos to the types.SignalInfo without soft deviation
// Note: soft deviation can be use in the future to adjust the deviation
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
