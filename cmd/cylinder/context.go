package main

import (
	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

type Context struct {
	config  *cylinder.Config
	keyring keyring.Keyring
	home    string
}
