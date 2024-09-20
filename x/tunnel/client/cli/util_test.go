package cli

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSignalDeviations(t *testing.T) {
	// test case for valid signal deviations
	t.Run("valid signal deviations", func(t *testing.T) {
		signalDeviations := []SignalDeviation{
			{SignalID: "BTC", DeviationBPS: 2000},
			{SignalID: "ETH", DeviationBPS: 4000},
		}
		file, cleanup := createTempSignalDeviationFile(signalDeviations)
		defer cleanup()

		result, err := parseSignalDeviations(file)

		require.NoError(t, err)
		require.Equal(t, signalDeviations, result.SignalDeviations)
	})

	// test case for empty file path
	t.Run("empty file path", func(t *testing.T) {
		result, err := parseSignalDeviations("")

		require.NoError(t, err)
		require.Equal(t, SignalDeviations{}, result)
	})
}

// Helper function to create a temporary file with signal info JSON content
func createTempSignalDeviationFile(signalDeviations []SignalDeviation) (string, func()) {
	file, err := os.CreateTemp("", "signalDeviations*.json")
	if err != nil {
		panic(err)
	}
	filePath := file.Name()

	data := struct {
		SignalDeviations []SignalDeviation `json:"signal_deviations"`
	}{SignalDeviations: signalDeviations}

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
