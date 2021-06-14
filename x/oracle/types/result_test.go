package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Using for evm test to show how to encode result
func TestEncodeResult(t *testing.T) {
	result := NewResult(
		"beeb",
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
	expectedEncodedResult := mustDecodeString("0a046265656210011a0f0000000342544300000000000003e8200128013002380140d8f7f8f60548daf7f8f60550015a0800000000009443ee")
	require.Equal(t, expectedEncodedResult, ModuleCdc.MustMarshalBinaryBare(&result))
}

func TestEncodeResultOfEmptyClientID(t *testing.T) {
	result := NewResult(
		"",
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
	expectedEncodedResult := mustDecodeString("10011a0f0000000342544300000000000003e82001280130013801409af6f8f605489df6f8f60550015a080000000000944387")
	require.Equal(t, expectedEncodedResult, ModuleCdc.MustMarshalBinaryBare(&result))
}

func TestEncodeFailResult(t *testing.T) {
	result := NewResult(
		"client_id",
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
	expectedEncodedResult := mustDecodeString("0a09636c69656e745f696410011a0f0000000342544300000000000003e82001280130013801409af6f8f605489df6f8f6055002")
	require.Equal(t, expectedEncodedResult, ModuleCdc.MustMarshalBinaryBare(&result))
}
