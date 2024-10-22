package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	sdk "github.com/cosmos/cosmos-sdk/types"

	app "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/cmd/bandd/cmd"
)

func main() {
	app.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(sdk.GetConfig())
	rootCmd := cmd.NewRootCmd()

	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
