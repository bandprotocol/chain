package yoda

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	band "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/app/params"
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
	NodeURI           string `mapstructure:"node"`                // Remote RPC URI of BandChain node to connect to
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
	cfg             Config
	kb              keyring.Keyring
	DefaultYodaHome string
)

func initConfig(c *Context) error {
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

	DefaultYodaHome = filepath.Join(userHomeDir, ".yoda")
}

func Main() {
	appConfig := sdk.GetConfig()
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(appConfig)

	ctx := &Context{}
	rootCmd := &cobra.Command{
		Use:   "yoda",
		Short: "BandChain oracle daemon to subscribe and response to oracle requests",
	}

	rootCmd.AddCommand(
		configCmd(),
		keysCmd(),
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

		tempDir := tempDir()
		initAppOptions := viper.New()
		initAppOptions.Set(flags.FlagHome, tempDir)
		tempApplication := band.NewBandApp(
			log.NewNopLogger(),
			dbm.NewMemDB(),
			nil,
			true,
			map[int64]bool{},
			tempDir,
			initAppOptions,
			100,
		)
		defer func() {
			if err := tempApplication.Close(); err != nil {
				panic(err)
			}
			if tempDir != band.DefaultNodeHome {
				os.RemoveAll(tempDir)
			}
		}()

		ctx.encodingConfig = params.EncodingConfig{
			InterfaceRegistry: tempApplication.InterfaceRegistry(),
			Codec:             tempApplication.AppCodec(),
			TxConfig:          tempApplication.GetTxConfig(),
			Amino:             tempApplication.LegacyAmino(),
		}

		kb, err = keyring.New("band", keyring.BackendTest, home, nil, tempApplication.AppCodec())
		if err != nil {
			return err
		}
		return initConfig(ctx)
	}
	rootCmd.PersistentFlags().String(flags.FlagHome, DefaultYodaHome, "home directory")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var tempDir = func() string {
	dir, err := os.MkdirTemp("", ".band")
	if err != nil {
		dir = band.DefaultNodeHome
	}
	defer os.RemoveAll(dir)

	return dir
}
