package main

import (
	"fmt"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"

	band "github.com/bandprotocol/chain/v2/app"
)

func main() {
	appConfig := sdk.GetConfig()
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(appConfig)

	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
