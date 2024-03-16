package testdata

import (
	"os"
	"os/exec"

	owasm "github.com/bandprotocol/go-owasm/api"

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

func wat2wasm(wat []byte) []byte {
	inputFile, err := os.CreateTemp("", "input")
	if err != nil {
		panic(err)
	}
	defer os.Remove(inputFile.Name())

	outputFile, err := os.CreateTemp("", "output")
	if err != nil {
		panic(err)
	}
	defer os.Remove(outputFile.Name())

	if _, err := inputFile.Write(wat); err != nil {
		panic(err)
	}

	if err := exec.Command("wat2wasm", inputFile.Name(), "-o", outputFile.Name()).Run(); err != nil { //nolint:gosec
		panic(err)
	}

	output, err := os.ReadFile(outputFile.Name())
	if err != nil {
		panic(err)
	}

	return output
}
