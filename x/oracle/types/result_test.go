package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Using for evm test to show how to encode result
func TestEncodeResult(t *testing.T) {
	result := NewResult(
		"beeb",
		0,
		1,
		mustDecodeString("0000000342544300000000000003e8"),
		1,
		1,
		2,
		1,
		1591622616,
		1591622618,
		RESOLVE_STATUS_SUCCESS,
		mustDecodeString("00000000009443ee"),
	)
	expectedEncodedResult := mustDecodeString(
		"0a04626565621801220f0000000342544300000000000003e8280130013802400148d8f7f8f60550daf7f8f6055801620800000000009443ee",
	)
	require.Equal(t, expectedEncodedResult, ModuleCdc.MustMarshal(&result))
}

func TestEncodeResultOfEmptyClientID(t *testing.T) {
	result := NewResult(
		"",
		0,
		1,
		mustDecodeString("0000000342544300000000000003e8"),
		1,
		1,
		1,
		1,
		1591622426,
		1591622429,
		RESOLVE_STATUS_SUCCESS,
		mustDecodeString("0000000000944387"),
	)
	expectedEncodedResult := mustDecodeString(
		"1801220f0000000342544300000000000003e82801300138014001489af6f8f605509df6f8f605580162080000000000944387",
	)
	require.Equal(t, expectedEncodedResult, ModuleCdc.MustMarshal(&result))
}

func TestEncodeFailResult(t *testing.T) {
	result := NewResult(
		"client_id",
		0,
		1,
		mustDecodeString("0000000342544300000000000003e8"),
		1,
		1,
		1,
		1,
		1591622426,
		1591622429,
		RESOLVE_STATUS_FAILURE,
		[]byte{},
	)
	expectedEncodedResult := mustDecodeString(
		"0a09636c69656e745f69641801220f0000000342544300000000000003e82801300138014001489af6f8f605509df6f8f6055802",
	)
	require.Equal(t, expectedEncodedResult, ModuleCdc.MustMarshal(&result))
}
