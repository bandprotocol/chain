package main

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/bandprotocol/chain/v2/cylinder"
)

// Context represents the application context.
type Context struct {
	config  *cylinder.Config // Configuration for the application.
	keyring keyring.Keyring  // Keyring for key management.
	home    string           // Home directory for the application.
}
