package cli

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestParseSignalInfos(t *testing.T) {
	// Test case for valid signal info
	t.Run("valid signal info", func(t *testing.T) {
		// Setup
		signalInfos := []types.SignalInfo{
			{SignalID: "BTC", DeviationBPS: 10, Interval: 10},
			{SignalID: "ETH", DeviationBPS: 10, Interval: 10},
		}
		file, cleanup := createTempSignalInfoFile(signalInfos)
		defer cleanup()

		// Execute
		result, err := parseSignalInfos(file)

		// Verify
		require.NoError(t, err)
		require.Equal(t, signalInfos, result)
	})

	// Test case for empty file path
	t.Run("empty file path", func(t *testing.T) {
		result, err := parseSignalInfos("")

		require.NoError(t, err)
		require.Nil(t, result)
	})
}

// Helper function to create a temporary file with signal info JSON content
func createTempSignalInfoFile(signalInfos []types.SignalInfo) (string, func()) {
	file, err := os.CreateTemp("", "signalInfos*.json")
	if err != nil {
		panic(err)
	}
	filePath := file.Name()

	data := struct {
		SignalInfos []types.SignalInfo `json:"signal_infos"`
	}{SignalInfos: signalInfos}

	content, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	if _, err := file.Write(content); err != nil {
		panic(err)
	}
	if err := file.Close(); err != nil {
		panic(err)
	}

	return filePath, func() { os.Remove(filePath) }
}
