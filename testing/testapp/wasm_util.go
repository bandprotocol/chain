package testapp

import (
	"os"
	"os/exec"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

func compile(code []byte) []byte {
	compiled, err := OwasmVM.Compile(code, types.MaxCompiledWasmCodeSize)
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
	if err := exec.Command("wat2wasm", inputFile.Name(), "-o", outputFile.Name()).Run(); err != nil {
		panic(err)
	}
	output, err := os.ReadFile(outputFile.Name())
	if err != nil {
		panic(err)
	}
	return output
}
