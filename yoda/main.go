package yoda

import (
	"fmt"
	odin "github.com/GeoDB-Limited/odin-core/app"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"
	"os"
)

const (
	flagValidator        = "validator"
	flagLogLevel         = "log-level"
	flagExecutor         = "executor"
	flagBroadcastTimeout = "broadcast-timeout"
	flagRPCPollInterval  = "rpc-poll-interval"
	flagMaxTry           = "max-try"
	flagMaxReport        = "max-report"
)

const (
	DefaultKeyringBackend = "test"
	DefaultHomeEnv        = "$HOME/.faucet"
)

// Global instance.
var (
	yoda Yoda
)

type Yoda struct {
	config  Config
	keybase keyring.Keyring
}

func Main() {
	appConfig := sdk.GetConfig()
	odin.SetBech32AddressPrefixesAndBip44CoinType(appConfig)
	appConfig.Seal()

	ctx := &Context{}
	rootCmd := &cobra.Command{
		Use:   "yoda",
		Short: "Odin oracle daemon to subscribe and response to oracle requests",
	}

	rootCmd.AddCommand(
		configCmd(),
		keysCmd(),
		runCmd(ctx),
	)
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		home, err := rootCmd.PersistentFlags().GetString(flags.FlagHome)
		if err != nil {
			return sdkerrors.Wrap(err, "failed to parse home directory flag")
		}
		keyringBackend, err := rootCmd.Flags().GetString(flags.FlagKeyringBackend)
		if err != nil {
			return sdkerrors.Wrap(err, "failed to parse keyring backend")
		}
		if err := os.MkdirAll(home, os.ModePerm); err != nil {
			return sdkerrors.Wrap(err, "failed to create a directory")
		}
		yoda.keybase, err = keyring.New(sdk.KeyringServiceName(), keyringBackend, home, nil)
		if err != nil {
			return sdkerrors.Wrap(err, "failed to create a new keyring")
		}
		return initConfig(home)
	}
	rootCmd.PersistentFlags().String(flags.FlagHome, os.ExpandEnv(DefaultHomeEnv), "home directory")
	rootCmd.PersistentFlags().String(flags.FlagKeyringBackend, DefaultKeyringBackend, "keyring backend")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
