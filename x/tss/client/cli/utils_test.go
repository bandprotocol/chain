package cli

import (
	"fmt"
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
				"key_sym": "A12yoSWiMwC+8k5XiD9UdQOrJZipntB9ZdSCtOof+O0m",
				"sig": "Aj1c3dvb5QNZAjHpqAljSM8n2TcUAh/q75GzwJVTcjujxdE324C0ZCgl5IxCVFDxRzHnzTwjl6u0sscOZacLBi4="
			}
		]
	}`
	_, err = tempFile.WriteString(validJSON)
	require.NoError(t, err)

	complains, err = parseComplains(tempFile.Name())
	require.NoError(t, err)
	require.Equal(t, 1, len(complains))

	fmt.Println(complains)

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
