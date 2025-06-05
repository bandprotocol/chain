//go:build !muslc

package testdata

import (
	owasm "github.com/bandprotocol/go-owasm/api"

	"github.com/bandprotocol/chain/v3/x/oracle/types"
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
