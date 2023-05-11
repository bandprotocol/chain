package main

import (
	"os"
	"path"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	band "github.com/bandprotocol/chain/v2/app"
)

// Global instances.
var (
	DefaultHome = filepath.Join(os.Getenv("HOME"), ".cylinder")
	cdc         = band.MakeEncodingConfig().Marshaler
)

func initConfig(ctx *Context, cmd *cobra.Command) error {
	if err := os.MkdirAll(ctx.home, os.ModePerm); err != nil {
		return err
	}

	var err error
	ctx.keyring, err = keyring.New("band", keyring.BackendTest, ctx.home, nil, cdc)
	if err != nil {
		return err
	}

	viper.SetConfigFile(path.Join(ctx.home, "config.yaml"))
	_ = viper.ReadInConfig() // If we fail to read config file, we'll just rely on cmd flags.

	if err := viper.Unmarshal(&ctx.config); err != nil {
		return err
	}

	return nil
}

func NewRootCmd() *cobra.Command {
	ctx := &Context{}
	rootCmd := &cobra.Command{
		Use:   "cylinder",
		Short: "BandChain oracle daemon to subscribe and response to signature requests",
	}

	rootCmd.AddCommand(
		configCmd(ctx),
		keysCmd(ctx),
		runCmd(ctx),
		version.NewVersionCommand(),
	)

	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(ctx, rootCmd)
	}

	rootCmd.PersistentFlags().StringVar(&ctx.home, flags.FlagHome, DefaultHome, "home directory")

	return rootCmd
}
