package main

import (
	rpcclient "github.com/cometbft/cometbft/rpc/client"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/app/params"
)

type Context struct {
	encodingConfig params.EncodingConfig
	client         rpcclient.Client
	gasPrices      sdk.DecCoins
	keys           chan keyring.Record
	amount         sdk.Coins
}
