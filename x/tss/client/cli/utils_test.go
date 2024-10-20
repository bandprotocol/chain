package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func TestParseComplaints(t *testing.T) {
	// 1. Test with empty string
	complaints, err := parseComplaints("")
	require.NoError(t, err)
	require.Empty(t, complaints)

	// 2. Test with a valid file
	// Write a valid JSON to a temp file
	tempFile, err := os.CreateTemp("", "complaints")

	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	validJSON := `{
		"complaints": [
			{
				"complainant": 1,
				"respondent": 2,
				"key_sym": "035db2a125a23300bef24e57883f547503ab2598a99ed07d65d482b4ea1ff8ed26",
				"signature": "023d5cdddbdbe503590231e9a8096348cf27d93714021feaef91b3c09553723ba3c5d137db80b4642825e48c425450f14731e7cd3c2397abb4b2c70e65a70b062e"
			}
		]
	}`
	_, err = tempFile.WriteString(validJSON)
	require.NoError(t, err)

	complaints, err = parseComplaints(tempFile.Name())
	require.NoError(t, err)
	require.Equal(t, 1, len(complaints))
	require.Equal(t, types.Complaint{
		Complainant: 1,
		Respondent:  2,
		KeySym:      testutil.HexDecode("035db2a125a23300bef24e57883f547503ab2598a99ed07d65d482b4ea1ff8ed26"),
		Signature: testutil.HexDecode(
			"023d5cdddbdbe503590231e9a8096348cf27d93714021feaef91b3c09553723ba3c5d137db80b4642825e48c425450f14731e7cd3c2397abb4b2c70e65a70b062e",
		),
	}, complaints[0])

	// 3. Test with a non-existent file
	complaints, err = parseComplaints("non-existent-file")
	require.Error(t, err)
	require.Nil(t, complaints)

	// 4. Test with an invalid JSON file
	invalidFile, err := os.CreateTemp("", "invalid-complaints")
	require.NoError(t, err)
	defer os.Remove(invalidFile.Name())

	invalidJSON := `[{invalidJSON}]`
	_, err = invalidFile.WriteString(invalidJSON)
	require.NoError(t, err)

	complaints, err = parseComplaints(invalidFile.Name())
	require.Error(t, err)
	require.Nil(t, complaints)
}
