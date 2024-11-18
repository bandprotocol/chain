package main

import (
	"fmt"
	"os"
	"path"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	band "github.com/bandprotocol/chain/v3/app"
)

const (
	flagPort   = "port"
	flagAmount = "amount"
)

// Config data structure for faucet server.
type Config struct {
	ChainID   string `mapstructure:"chain-id"`   // ChainID of the target chain
	NodeURI   string `mapstructure:"node"`       // Remote RPC URI of BandChain node to connect to
	GasPrices string `mapstructure:"gas-prices"` // Gas prices of the transaction
	Port      string `mapstructure:"port"`       // Port of faucet service
	Amount    int64  `mapstructure:"amount"`     // Amount of BAND for each request
}

// Global instances.
var (
	cfg     Config
	keybase keyring.Keyring
)

func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(flags.FlagHome)
	if err != nil {
		return err
	}
	viper.SetConfigFile(path.Join(home, "config.yaml"))
	_ = viper.ReadInConfig() // If we fail to read config file, we'll just rely on cmd flags.
	if err := viper.Unmarshal(&cfg); err != nil {
		return err
	}
	return nil
}

func main() {
	appConfig := sdk.GetConfig()
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(appConfig)

	ctx := &Context{}
	rootCmd := &cobra.Command{
		Use:   "faucet",
		Short: "Faucet server for devnet",
	}

	rootCmd.AddCommand(
		configCmd(),
		keysCmd(ctx),
		runCmd(ctx),
	)
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		home, err := rootCmd.PersistentFlags().GetString(flags.FlagHome)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(home, os.ModePerm); err != nil {
			return err
		}

		initAppOptions := viper.New()
		tempDir := tempDir()
		initAppOptions.Set(flags.FlagHome, tempDir)
		tempApplication := band.NewBandApp(
			log.NewNopLogger(),
			dbm.NewMemDB(),
			nil,
			true,
			map[int64]bool{},
			tempDir,
			initAppOptions,
			[]wasmkeeper.Option{},
			100,
		)
		ctx.bandApp = tempApplication

		keybase, err = keyring.New("band", keyring.BackendTest, home, nil, tempApplication.AppCodec())
		if err != nil {
			return err
		}

		return initConfig(rootCmd)
	}
	rootCmd.PersistentFlags().String(flags.FlagHome, os.ExpandEnv("$HOME/.faucet"), "home directory")
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
