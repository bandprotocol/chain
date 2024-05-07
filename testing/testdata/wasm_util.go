package testdata

import (
	owasm "github.com/bandprotocol/go-owasm/api"
	"github.com/bytecodealliance/wasmtime-go/v20"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

func Compile(code []byte) []byte {
	owasmVM, err := owasm.NewVm(10)
	if err != nil {
		panic(err)
	}

	compiled, err := owasmVM.Compile(code, types.MaxCompiledWasmCodeSize)
	if err != nil {
		panic(err)
	}

	return compiled
}

func wat2wasm(wat string) []byte {
	wasm, err := wasmtime.Wat2Wasm(wat)
	if err != nil {
		panic(err)
	}

	return wasm
}
