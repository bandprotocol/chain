package main

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
	"github.com/bandprotocol/chain/v3/cmd/grogu/cmd"
	"github.com/bandprotocol/chain/v3/grogu/context"
	"github.com/bandprotocol/chain/v3/pkg/logger"
)

func main() {
	appConfig := sdk.GetConfig()
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(appConfig)

	ctx := &context.Context{}
	rootCmd := createRootCmd(ctx)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func createRootCmd(ctx *context.Context) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "grogu",
		Short: "BandChain daemon to submit signal prices for feeds module",
	}

	rootCmd.AddCommand(
		cmd.ConfigCmd(),
		cmd.KeysCmd(ctx),
		cmd.RunCmd(ctx),
		version.NewVersionCommand(),
	)

	rootCmd.PersistentPreRunE = createPersistentPreRunE(rootCmd, ctx)
	rootCmd.PersistentFlags().String(flags.FlagHome, getDefaultHome(), "home directory")
	return rootCmd
}

func createPersistentPreRunE(rootCmd *cobra.Command, ctx *context.Context) func(
	cmd *cobra.Command,
	args []string,
) error {
	return func(_ *cobra.Command, _ []string) error {
		home, err := rootCmd.PersistentFlags().GetString(flags.FlagHome)
		if err != nil {
			return err
		}

		if err = os.MkdirAll(home, os.ModePerm); err != nil {
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

		encodingConfig := params.EncodingConfig{
			InterfaceRegistry: tempApplication.InterfaceRegistry(),
			Codec:             tempApplication.AppCodec(),
			TxConfig:          tempApplication.GetTxConfig(),
			Amino:             tempApplication.LegacyAmino(),
		}

		kr, err := keyring.New("band", keyring.BackendTest, home, nil, tempApplication.AppCodec())
		if err != nil {
			return err
		}

		cfg, err := initConfig(home)
		if err != nil {
			return err
		}

		logger, err := initLogger(cfg.LogLevel)
		if err != nil {
			return err
		}

		*ctx = *context.New(*cfg, kr, logger, home, encodingConfig)
		return nil
	}
}

// initConfig initializes the configuration from viper config file/flag.
func initConfig(homePath string) (*context.Config, error) {
	viper.SetConfigFile(path.Join(homePath, "config.yaml"))

	// If the config file cannot be read, only cmd flags will be used.
	_ = viper.ReadInConfig()

	var cfg context.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// InitLog initializes the logger for the context.
func initLogger(logLevel string) (*logger.Logger, error) {
	allowLevel, err := log.ParseLogLevel(logLevel)
	if err != nil {
		return nil, err
	}

	return logger.NewLogger(allowLevel), nil
}

func getDefaultHome() string {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(userHomeDir, ".grogu")
}

var tempDir = func() string {
	dir, err := os.MkdirTemp("", ".band")
	if err != nil {
		dir = band.DefaultNodeHome
	}

	return dir
}
