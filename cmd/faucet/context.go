package main

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Context defines the structure of faucet context.
type Context struct {
	chainID                string
	nodeURI                string
	port                   string
	gasPrices              sdk.DecCoins
	coins                  sdk.Coins
	maxPerPeriodWithdrawal sdk.Coins
	keys                   chan keyring.Info
}

// initCtx parses config string literals.
func (ctx *Context) initCtx() error {
	var err error
	ctx.gasPrices, err = sdk.ParseDecCoins(faucet.config.GasPrices)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to parse gas prices")
	}
	ctx.coins, err = sdk.ParseCoinsNormalized(faucet.config.Coins)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to parse coins to withdraw")
	}
	ctx.maxPerPeriodWithdrawal, err = sdk.ParseCoinsNormalized(faucet.config.MaxPerPeriodWithdrawal)
	return sdkerrors.Wrap(err, "failed to parse max withdrawal per period")
}
