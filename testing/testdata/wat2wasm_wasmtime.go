//go:build !muslc
// +build !muslc

package testdata

import (
	"github.com/bytecodealliance/wasmtime-go/v20"
)

func wat2wasm(wat string) []byte {
	wasm, err := wasmtime.Wat2Wasm(wat)
	if err != nil {
		panic(err)
	}

	return wasm
}
