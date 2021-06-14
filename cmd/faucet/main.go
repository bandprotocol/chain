package main

import (
	"fmt"
	band "github.com/GeoDB-Limited/odin-core/app"
	"github.com/GeoDB-Limited/odin-core/cmd/faucet/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"os"
)

const (
	DefaultKeyringBackend = "test"
	DefaultHomeEnv        = "$HOME/.faucet"
)

func main() {
	appConfig := sdk.GetConfig()
	band.SetBech32AddressPrefixesAndBip44CoinType(appConfig)
	appConfig.Seal()

	cfg := &config.Config{}

	rootCmd := &cobra.Command{
		Use:   "faucet",
		Short: "Faucet server for developers' network",
	}
	rootCmd.AddCommand(
		runCmd(cfg),
		config.SetParamCmd(),
		config.KeysCmd(cfg),
	)
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		home, err := rootCmd.PersistentFlags().GetString(flags.FlagHome)
		if err != nil {
			return err
		}
		keyringBackend, err := rootCmd.Flags().GetString(flags.FlagKeyringBackend)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(home, os.ModePerm); err != nil {
			return err
		}
		cfg.Keyring, err = keyring.New(sdk.KeyringServiceName(), keyringBackend, home, nil)
		if err != nil {
			return err
		}
		return config.InitConfig(rootCmd, cfg)
	}
	rootCmd.PersistentFlags().String(flags.FlagHome, os.ExpandEnv(DefaultHomeEnv), "home directory")
	rootCmd.PersistentFlags().String(flags.FlagKeyringBackend, DefaultKeyringBackend, "keyring backend")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
