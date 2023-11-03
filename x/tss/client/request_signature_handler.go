package client

import (
	"github.com/spf13/cobra"
)

// function to create the cli handler
type CLIHandlerFn func() *cobra.Command

// Type for a RequestingSignatureHandler handler
type RequestingSignatureHandler struct {
	CLIHandler CLIHandlerFn
}

// NewRequestingSignatureHandler creates a new RequestingSignatureHandler object
func NewRequestingSignatureHandler(cliHandler CLIHandlerFn) RequestingSignatureHandler {
	return RequestingSignatureHandler{
		CLIHandler: cliHandler,
	}
}
