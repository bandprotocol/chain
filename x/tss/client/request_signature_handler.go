package client

import "github.com/spf13/cobra"

// function to create the cli handler
type CLIHandlerFn func() *cobra.Command

// Type for a RequestSignatureHandler handler
type RequestSignatureHandler struct {
	CLIHandler CLIHandlerFn
}

// RequestSignatureHandler creates a new RequestSignatureHandler object
func NewRequestSignatureHandler(cliHandler CLIHandlerFn) RequestSignatureHandler {
	return RequestSignatureHandler{
		CLIHandler: cliHandler,
	}
}
