package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"path"
	"time"
)

const (
	ConfigurationFile = "config.yaml"
)

// Config data structure for faucet server.
type Config struct {
	ChainID                string        `mapstructure:"chain-id"`                  // ChainID of the target chain
	NodeURI                string        `mapstructure:"node"`                      // Remote RPC URI of BandChain node to connect to
	GasPrices              string        `mapstructure:"gas-prices"`                // Gas prices of the transaction
	Port                   string        `mapstructure:"port"`                      // Port of faucet service
	Coins                  string        `mapstructure:"coins"`                     // Coins is amount of coins to withdraw
	Period                 time.Duration `mapstructure:"period"`                    // Period is period of withdrawal limitation
	MaxPerPeriodWithdrawal string        `mapstructrue:"max-per-period-withdrawal"` // MaxPerPeriodWithdrawal is max amount to withdraw
}

// configCmd sets a new config parameter.
func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config [key] [value]",
		Aliases: []string{"c"},
		Short:   "Set faucet configuration environment",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.Set(args[0], args[1])
			return viper.WriteConfig()
		},
	}
	return cmd
}

// initConfig initializes faucet config from env.
func initConfig(homePath string) error {
	viper.SetConfigFile(path.Join(homePath, ConfigurationFile))
	_ = viper.ReadInConfig() // If we fail to read config file, we'll just rely on cmd flags.
	if err := viper.Unmarshal(&faucet.config); err != nil {
		return err
	}
	return nil
}
