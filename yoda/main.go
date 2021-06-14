package yoda

import (
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"os"
	"path"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	odin "github.com/GeoDB-Limited/odin-core/app"
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

// Config data structure for yoda daemon.
type Config struct {
	ChainID           string `mapstructure:"chain-id"`            // ChainID of the target chain
	NodeURI           string `mapstructure:"node"`                // Remote RPC URI of Odin node to connect to
	Validator         string `mapstructure:"validator"`           // The validator address that I'm responsible for
	GasPrices         string `mapstructure:"gas-prices"`          // Gas prices of the transaction
	LogLevel          string `mapstructure:"log-level"`           // Log level of the logger
	Executor          string `mapstructure:"executor"`            // Executor name and URL (example: "Executor name:URL")
	BroadcastTimeout  string `mapstructure:"broadcast-timeout"`   // The time that Yoda will wait for tx commit
	RPCPollInterval   string `mapstructure:"rpc-poll-interval"`   // The duration of rpc poll interval
	MaxTry            uint64 `mapstructure:"max-try"`             // The maximum number of tries to submit a report transaction
	MaxReport         uint64 `mapstructure:"max-report"`          // The maximum number of reports in one transaction
	MetricsListenAddr string `mapstructure:"metrics-listen-addr"` // Address to listen on for prometheus metrics
}

// Global instances.
var (
	yoda Yoda
)

type Yoda struct {
	config  Config
	keybase keyring.Keyring
}


func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(flags.FlagHome)
	if err != nil {
		return err
	}
	viper.SetConfigFile(path.Join(home, "config.yaml"))
	_ = viper.ReadInConfig() // If we fail to read config file, we'll just rely on cmd flags.
	if err := viper.Unmarshal(&yoda.config); err != nil {
		return sdkerrors.Wrap(err, "failed to unmarshal config")
	}
	return nil
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
		if err := os.MkdirAll(home, os.ModePerm); err != nil {
			return err
		}
		yoda.keybase, err = keyring.New("odin", "test", home, nil)
		if err != nil {
			return sdkerrors.Wrap(err, "failed to create a new keyring")
		}
		return initConfig(rootCmd)
	}
	rootCmd.PersistentFlags().String(flags.FlagHome, os.ExpandEnv("$HOME/.yoda"), "home directory")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
