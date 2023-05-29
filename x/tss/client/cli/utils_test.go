package cli

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseComplains(t *testing.T) {
	// 1. Test with empty string
	complains, err := parseComplains("")
	require.NoError(t, err)
	require.Empty(t, complains)

	// 2. Test with a valid file
	// Write a valid JSON to a temp file
	tempFile, err := ioutil.TempFile("", "complains")

	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	validJSON := `{
		"complains": [
			{
				"i": 1,
				"j": 2,
				"key_sym": "a2V5X3N5bQ==",
				"signature": "c2lnbmF0dXJl",
				"nonce_sym": "bm9uY2Vfc3lt"
			}
		]
	}`
	_, err = tempFile.WriteString(validJSON)
	require.NoError(t, err)

	complains, err = parseComplains(tempFile.Name())
	require.NoError(t, err)
	require.Equal(t, 1, len(complains))

	// 3. Test with a non-existent file
	complains, err = parseComplains("non-existent-file")
	require.Error(t, err)
	require.Nil(t, complains)

	// 4. Test with an invalid JSON file
	invalidFile, err := ioutil.TempFile("", "invalidComplains")
	require.NoError(t, err)
	defer os.Remove(invalidFile.Name())

	invalidJSON := `[{invalidJSON}]`
	_, err = invalidFile.WriteString(invalidJSON)
	require.NoError(t, err)

	complains, err = parseComplains(invalidFile.Name())
	require.Error(t, err)
	require.Nil(t, complains)
}
