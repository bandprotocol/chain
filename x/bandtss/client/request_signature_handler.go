package client

import (
	"github.com/spf13/cobra"
)

// function to create the cli handler
type CLIHandlerFn func() *cobra.Command

// RequestSignatureHandler wraps CLIHandlerFn
type RequestSignatureHandler struct {
	CLIHandler CLIHandlerFn
}

// NewRequestSignatureHandler creates a new SignatureOrderHandler object
func NewRequestSignatureHandler(cliHandler CLIHandlerFn) RequestSignatureHandler {
	return RequestSignatureHandler{
		CLIHandler: cliHandler,
	}
}
