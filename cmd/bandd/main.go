package main

import (
	"os"

	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	sdk "github.com/cosmos/cosmos-sdk/types"

	app "github.com/bandprotocol/chain/v2/app"
	"github.com/bandprotocol/chain/v2/cmd/bandd/cmd"
)

func main() {
	app.SetBech32AddressPrefixesAndBip44CoinType(sdk.GetConfig())
	rootCmd, _ := cmd.NewRootCmd()

	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		switch e := err.(type) {
		case server.ErrorCode:
			os.Exit(e.Code)

		default:
			os.Exit(1)
		}
	}
}
