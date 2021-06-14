package config

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"path"
	"time"
)

const (
	ConfigurationFile = "config.yaml"

	ChainIDParam                = "chain-id"
	NodeURIParam                = "node"
	GasPricesParam              = "gas-prices"
	PortParam                   = "port"
	CoinsParam                  = "coins"
	PeriodParam                 = "period"
	MaxWithdrawalPerPeriodParam = "max-withdrawal-per-period"
)

// Config data structure for faucet server.
type Config struct {
	ChainID                string          // ChainID of the target chain
	NodeURI                string          // Remote RPC URI of BandChain node to connect to
	Port                   string          // Port of faucet service
	GasPrices              sdk.DecCoins    // Gas prices of the transaction
	Period                 time.Duration   // Period is period of withdrawal limitation
	MaxWithdrawalPerPeriod sdk.Coins       // MaxWithdrawalPerPeriod is max amount to withdraw
	Coins                  sdk.Coins       // Coins is amount of coins to withdraw
	Keyring                keyring.Keyring // Keyring stores the accounts that provide funds for faucet
}

// InitConfig initializes config from viper
func InitConfig(cmd *cobra.Command, cfg *Config) error {
	home, err := cmd.PersistentFlags().GetString(flags.FlagHome)
	if err != nil {
		return err
	}
	viper.SetConfigFile(path.Join(home, ConfigurationFile))
	_ = viper.ReadInConfig() // If we fail to read config file, we'll just rely on cmd flags.

	cfg.ChainID = viper.GetString(ChainIDParam)
	cfg.NodeURI = viper.GetString(NodeURIParam)
	cfg.Port = viper.GetString(PortParam)
	cfg.Period, err = time.ParseDuration(viper.GetString(PeriodParam))
	if err != nil {
		return err
	}
	cfg.GasPrices, err = sdk.ParseDecCoins(viper.GetString(GasPricesParam))
	if err != nil {
		return err
	}
	cfg.Coins, err = sdk.ParseCoinsNormalized(viper.GetString(CoinsParam))
	if err != nil {
		return err
	}
	cfg.MaxWithdrawalPerPeriod, err = sdk.ParseCoinsNormalized(viper.GetString(MaxWithdrawalPerPeriodParam))
	if err != nil {
		return err
	}

	return nil
}

// SetParamCmd sets a new config parameter.
func SetParamCmd() *cobra.Command {
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
