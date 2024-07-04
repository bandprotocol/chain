package cli

import (
	"encoding/json"
	"os"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// parseSignalInfos parses the signal infos from the given file
func parseSignalInfos(signalInfosFile string) ([]types.SignalInfo, error) {
	signalInfos := struct {
		SignalInfos []types.SignalInfo `json:"signal_infos"`
	}{}

	if signalInfosFile == "" {
		return signalInfos.SignalInfos, nil
	}

	contents, err := os.ReadFile(signalInfosFile)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(contents, &signalInfos); err != nil {
		return nil, err
	}

	return signalInfos.SignalInfos, nil
}
