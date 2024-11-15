package types

import "fmt"

// AbsInt64 returns an absolute of int64.
// Panics on min int64 (-9223372036854775808).
func AbsInt64(x int64) int64 {
	if x < 0 {
		return -1 * x
	}

	return x
}

// stringToBytes32 converts a string to a fixed size byte array.
func stringToBytes32(str string) ([32]byte, error) {
	if len(str) > 32 {
		return [32]byte{}, fmt.Errorf("string is too long")
	}

	var byteArray [32]byte
	copy(byteArray[32-len(str):], str)
	return byteArray, nil
}
