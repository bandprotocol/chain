package schnorr

import (
	"encoding/hex"
	"errors"
	"testing"
)

// TestSignatureParsing ensures that signatures are properly parsed including
// error paths.
func TestSignatureParsing(t *testing.T) {
	tests := []struct {
		name string // test description
		sig  string // hex encoded signature to parse
		err  error  // expected error
	}{{
		name: "valid signature 1",
		sig: "02c6ec70969d8367538c442f8e13eb20ff0c9143690f31cd3a384da54dd29ec0aa" +
			"4b78a1b0d6b4186195d42a85614d3befd9f12ed26542d0dd1045f38c98b4a405",
		err: nil,
	}, {
		name: "valid signature 2",
		sig: "02adc21db084fa1765f9372c2021fb298720f3d13e6d844e2dff751a2d46a69277" +
			"0b989e316f7faf308a5f4a7343c0569465287cf6bff457250d6dacbb361f6e63",
		err: nil,
	}, {
		name: "empty",
		sig:  "",
		err:  ErrSigTooShort,
	}, {
		name: "too short by one byte",
		sig: "adc21db084fa1765f9372c2021fb298720f3d13e6d844e2dff751a2d46a69277" +
			"0b989e316f7faf308a5f4a7343c0569465287cf6bff457250d6dacbb361f6e",
		err: ErrSigTooShort,
	}, {
		name: "too long by one byte",
		sig: "02adc21db084fa1765f9372c2021fb298720f3d13e6d844e2dff751a2d46a69277" +
			"0b989e316f7faf308a5f4a7343c0569465287cf6bff457250d6dacbb361f6e6300",
		err: ErrSigTooLong,
	}, {
		name: "r == p",
		sig: "02fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f" +
			"181522ec8eca07de4860a4acdd12909d831cc56cbbac4622082221a8768d1d09",
		err: ErrSigRTooBig,
	}, {
		name: "r > p",
		sig: "02fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc30" +
			"181522ec8eca07de4860a4acdd12909d831cc56cbbac4622082221a8768d1d09",
		err: ErrSigRTooBig,
	}, {
		name: "s == n",
		sig: "024e45e16932b8af514961a1d3a1a25fdf3f4f7732e9d624c6c61548ab5fb8cd41" +
			"fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141",
		err: ErrSigSTooBig,
	}, {
		name: "s > n",
		sig: "024e45e16932b8af514961a1d3a1a25fdf3f4f7732e9d624c6c61548ab5fb8cd41" +
			"fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364142",
		err: ErrSigSTooBig,
	}}

	for _, test := range tests {
		_, err := ParseSignature(hexToBytes(test.sig))
		if !errors.Is(err, test.err) {
			t.Errorf("%s mismatched err -- got %v, want %v", test.name, err,
				test.err)
			continue
		}
	}
}

// hexToBytes converts the passed hex string into bytes and will panic if there
// is an error.  This is only provided for the hard-coded constants so errors in
// the source code can be detected. It will only (and must only) be called with
// hard-coded values.
func hexToBytes(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("invalid hex in source file: " + s)
	}
	return b
}
