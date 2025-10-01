package cli

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSignalDeviations(t *testing.T) {
	signalDeviations := []SignalDeviation{
		{SignalID: "CS:BTC-USD", DeviationBPS: 2000},
		{SignalID: "CS:ETH-USD", DeviationBPS: 4000},
	}
	file, cleanup := createTempSignalDeviationFile(signalDeviations)
	defer cleanup()

	result, err := parseSignalDeviations(file)
	require.NoError(t, err)
	require.Equal(t, signalDeviations, result.SignalDeviations)
}

// createTempSignalDeviationFile is a helper function to create a temporary file with signal info JSON content
func createTempSignalDeviationFile(signalDeviations []SignalDeviation) (string, func()) {
	file, err := os.CreateTemp("", "signalDeviations*.json")
	if err != nil {
		panic(err)
	}
	filePath := file.Name()

	content, err := json.Marshal(SignalDeviations{SignalDeviations: signalDeviations})
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
