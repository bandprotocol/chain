package crypto

import (
	"encoding/hex"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// hexToBytes converts the passed hex string into bytes and will panic if there
// is an error.  This is only provided for the hard-coded constants so errors in
// the source code can be detected. It will only (and must only) be called with
// hard-coded values.
func HexToBytes(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("invalid hex in source file: " + s)
	}
	return b
}

// hexToFieldVal converts the passed hex string into a FieldVal and will panic
// if there is an error.  This is only provided for the hard-coded constants so
// errors in the source code can be detected. It will only (and must only) be
// called with hard-coded values.
func HexToFieldVal(s string) *secp256k1.FieldVal {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("invalid hex in source file: " + s)
	}
	var f secp256k1.FieldVal
	if overflow := f.SetByteSlice(b); overflow {
		panic("hex in source file overflows mod P: " + s)
	}
	return &f
}

func BytesToFieldVal(b []byte) secp256k1.FieldVal {
	var f secp256k1.FieldVal
	if overflow := f.SetByteSlice(b); overflow {
		panic("byte in source file overflows")
	}
	return f
}

// jacobianPointFromHex decodes the passed big-endian hex strings into a
// Jacobian point with its internal fields set to the resulting values.  Only
// the first 32-bytes are used.
func JacobianPointFromHex(x, y string) secp256k1.JacobianPoint {
	var p secp256k1.JacobianPoint
	p.X.Set(HexToFieldVal(x))
	p.Y.Set(HexToFieldVal(y))
	p.Z.Set(HexToFieldVal("1"))
	return p
}
