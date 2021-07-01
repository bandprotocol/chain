package yoda

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"path"
)

const (
	ConfigurationFile = "config.yaml"
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

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config [key] [value]",
		Aliases: []string{"c"},
		Short:   "Set yoda configuration environment",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.Set(args[0], args[1])
			return viper.WriteConfig()
		},
	}
	return cmd
}

func initConfig(homePath string) error {
	viper.SetConfigFile(path.Join(homePath, ConfigurationFile))
	_ = viper.ReadInConfig() // If we fail to read config file, we'll just rely on cmd flags.
	if err := viper.Unmarshal(&yoda.config); err != nil {
		return sdkerrors.Wrap(err, "failed to unmarshal config")
	}
	return nil
}
