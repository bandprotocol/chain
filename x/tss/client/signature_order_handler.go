package client

import (
	"github.com/spf13/cobra"
)

// function to create the cli handler
type CLIHandlerFn func() *cobra.Command

// SignatureOrderHandler wraps CLIHandlerFn
type SignatureOrderHandler struct {
	CLIHandler CLIHandlerFn
}

// NewSignatureOrderHandler creates a new SignatureOrderHandler object
func NewSignatureOrderHandler(cliHandler CLIHandlerFn) SignatureOrderHandler {
	return SignatureOrderHandler{
		CLIHandler: cliHandler,
	}
}
