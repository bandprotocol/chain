package main

import (
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Context struct {
	client    rpcclient.Client
	gasPrices sdk.DecCoins
	keys      chan keyring.Record
	amount    sdk.Coins
}
