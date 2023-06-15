package schnorr

import (
	"errors"
	"testing"
)

// TestComplaintSignatureParsing ensures that complaint signatures are properly parsed including
// error paths.
func TestComplaintSignatureParsing(t *testing.T) {
	tests := []struct {
		name string // test description
		sig  string // hex encoded signature to parse
		err  error  // expected error
	}{{
		name: "valid signature 1",
		sig:  "02a55f7d417d1b51d91e6097473f00f528291aaa0dd11733e83eb85680ed5d4e36034946dba60574e576aef1c252e48db7c2c40f828efdb374ec8bd48ea36af06ac89fe3b8aef036713c547118f5a0adb8108dfe19b4067081f26a2fe27a87f60c0b",
		err:  nil,
	}, {
		name: "valid signature 2",
		sig:  "03000b8376dd57146397b8b38edd9f4bd551dfd3d6955c1dbdad1115da1a8fcaf103bb20bf99b70ae76cf2ef8779d0d88f8bb3eada6dd25f1663738f290ce9595b11110db1b2cbfc92e84076de48b1636a480fefcb1df6a4bdc4cea33d45b1851631",
		err:  nil,
	}, {
		name: "empty",
		sig:  "",
		err:  ErrSigTooShort,
	}, {
		name: "too short by one byte",
		sig:  "000b8376dd57146397b8b38edd9f4bd551dfd3d6955c1dbdad1115da1a8fcaf103bb20bf99b70ae76cf2ef8779d0d88f8bb3eada6dd25f1663738f290ce9595b11110db1b2cbfc92e84076de48b1636a480fefcb1df6a4bdc4cea33d45b1851631",
		err:  ErrSigTooShort,
	}, {
		name: "too long by one byte",
		sig:  "0203000b8376dd57146397b8b38edd9f4bd551dfd3d6955c1dbdad1115da1a8fcaf103bb20bf99b70ae76cf2ef8779d0d88f8bb3eada6dd25f1663738f290ce9595b11110db1b2cbfc92e84076de48b1636a480fefcb1df6a4bdc4cea33d45b1851631",
		err:  ErrSigTooLong,
	}, {
		name: "a1 == p",
		sig: "02fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f" +
			"03000b8376dd57146397b8b38edd9f4bd551dfd3d6955c1dbdad1115da1a8fcaf1" +
			"181522ec8eca07de4860a4acdd12909d831cc56cbbac4622082221a8768d1d09",
		err: ErrSigA1TooBig,
	}, {
		name: "a2 == p",
		sig: "03000b8376dd57146397b8b38edd9f4bd551dfd3d6955c1dbdad1115da1a8fcaf1" +
			"02fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f" +
			"181522ec8eca07de4860a4acdd12909d831cc56cbbac4622082221a8768d1d09",
		err: ErrSigA2TooBig,
	}, {
		name: "a1 > p",
		sig: "02fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc30" +
			"03000b8376dd57146397b8b38edd9f4bd551dfd3d6955c1dbdad1115da1a8fcaf1" +
			"181522ec8eca07de4860a4acdd12909d831cc56cbbac4622082221a8768d1d09",
		err: ErrSigA1TooBig,
	}, {
		name: "a2 > p",
		sig: "03000b8376dd57146397b8b38edd9f4bd551dfd3d6955c1dbdad1115da1a8fcaf1" +
			"02fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc30" +
			"181522ec8eca07de4860a4acdd12909d831cc56cbbac4622082221a8768d1d09",
		err: ErrSigA2TooBig,
	}, {
		name: "z == n",
		sig: "024e45e16932b8af514961a1d3a1a25fdf3f4f7732e9d624c6c61548ab5fb8cd41" +
			"024e45e16932b8af514961a1d3a1a25fdf3f4f7732e9d624c6c61548ab5fb8cd41" +
			"fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141",
		err: ErrSigZTooBig,
	}, {
		name: "z > n",
		sig: "024e45e16932b8af514961a1d3a1a25fdf3f4f7732e9d624c6c61548ab5fb8cd41" +
			"024e45e16932b8af514961a1d3a1a25fdf3f4f7732e9d624c6c61548ab5fb8cd41" +
			"fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364142",
		err: ErrSigZTooBig,
	}}

	for _, test := range tests {
		_, err := ParseComplaintSignature(hexToBytes(test.sig))
		if !errors.Is(err, test.err) {
			t.Errorf("%s mismatched err -- got %v, want %v", test.name, err,
				test.err)
			continue
		}
	}
}
