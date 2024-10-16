package main

import (
	"fmt"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"

	band "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/cylinder/context"
)

func main() {
	appConfig := sdk.GetConfig()
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(appConfig)

	ctx := &context.Context{}
	rootCmd := NewRootCmd(ctx)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
