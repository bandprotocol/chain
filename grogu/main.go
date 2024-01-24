package grogu

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	band "github.com/bandprotocol/chain/v2/app"
)

const (
	flagValidator        = "validator"
	flagLogLevel         = "log-level"
	flagPriceService     = "price-service"
	flagBroadcastTimeout = "broadcast-timeout"
	flagRPCPollInterval  = "rpc-poll-interval"
	flagMaxTry           = "max-try"
)

// Config data structure for grogu daemon.
type Config struct {
	ChainID          string `mapstructure:"chain-id"`          // ChainID of the target chain
	NodeURI          string `mapstructure:"node"`              // Remote RPC URI of BandChain node to connect to
	Validator        string `mapstructure:"validator"`         // The validator address that I'm responsible for
	GasPrices        string `mapstructure:"gas-prices"`        // Gas prices of the transaction
	LogLevel         string `mapstructure:"log-level"`         // Log level of the logger
	PriceService     string `mapstructure:"price-service"`     // PriceService name and URL (example: "PriceService name:URL")
	BroadcastTimeout string `mapstructure:"broadcast-timeout"` // The time that Grogu will wait for tx commit
	RPCPollInterval  string `mapstructure:"rpc-poll-interval"` // The duration of rpc poll interval
	MaxTry           uint64 `mapstructure:"max-try"`           // The maximum number of tries to submit a report transaction
}

// Global instances.
var (
	cfg              Config
	kb               keyring.Keyring
	DefaultGroguHome string
)

func initConfig(c *Context, cmd *cobra.Command) error {
	viper.SetConfigFile(path.Join(c.home, "config.yaml"))
	_ = viper.ReadInConfig() // If we fail to read config file, we'll just rely on cmd flags.
	if err := viper.Unmarshal(&cfg); err != nil {
		return err
	}
	return nil
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultGroguHome = filepath.Join(userHomeDir, ".grogu")
}

func Main() {
	appConfig := sdk.GetConfig()
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(appConfig)

	ctx := &Context{}
	rootCmd := &cobra.Command{
		Use:   "grogu",
		Short: "BandChain daemon to submit prices for feeds module",
	}

	rootCmd.AddCommand(
		configCmd(),
		keysCmd(ctx),
		runCmd(ctx),
		version.NewVersionCommand(),
	)
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		home, err := rootCmd.PersistentFlags().GetString(flags.FlagHome)
		if err != nil {
			return err
		}
		ctx.home = home
		if err := os.MkdirAll(home, os.ModePerm); err != nil {
			return err
		}
		kb, err = keyring.New("band", keyring.BackendTest, home, nil, cdc)
		if err != nil {
			return err
		}
		return initConfig(ctx, rootCmd)
	}
	rootCmd.PersistentFlags().String(flags.FlagHome, DefaultGroguHome, "home directory")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
