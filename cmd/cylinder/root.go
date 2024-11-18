package main

import (
	"encoding/hex"
	"os"
	"path"
	"path/filepath"
	"reflect"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/version"

	band "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/cylinder/context"
	"github.com/bandprotocol/chain/v3/pkg/tss"
)

func hexByteToScalarHookFunc() mapstructure.DecodeHookFunc {
	return func(
		from reflect.Type, // data type
		to reflect.Type, // target data type
		data interface{}, // raw data
	) (interface{}, error) {
		// Check if the data type matches the expected one
		if from.Kind() != reflect.String {
			return data, nil
		}

		// Check if the target type matches the expected one
		if to != reflect.TypeOf(tss.Scalar{}) {
			return data, nil
		}

		return hex.DecodeString(data.(string))
	}
}

// initConfig initializes the configuration.
func initConfig(ctx *context.Context, _ *cobra.Command) error {
	viper.SetConfigFile(path.Join(ctx.Home, "config.yaml"))
	_ = viper.ReadInConfig() // If we fail to read config file, we'll just rely on cmd flags.

	configOption := viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		hexByteToScalarHookFunc(),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	))

	if err := viper.Unmarshal(&ctx.Config, configOption); err != nil {
		return err
	}

	return ctx.InitLog()
}

// NewRootCmd returns a new instance of the root command.
func NewRootCmd(ctx *context.Context) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "cylinder",
		Short: "BandChain oracle daemon to subscribe and response to signature requests",
	}

	rootCmd.AddCommand(
		configCmd(ctx),
		keysCmd(ctx),
		importCmd(ctx),
		exportCmd(ctx),
		runCmd(ctx),
		version.NewVersionCommand(),
	)

	rootCmd.PersistentPreRunE = createPersistentPreRunE(rootCmd, ctx)
	rootCmd.PersistentFlags().StringVar(&ctx.Home, flags.FlagHome, getDefaultHome(), "home directory")

	return rootCmd
}

func createPersistentPreRunE(rootCmd *cobra.Command, ctx *context.Context) func(
	cmd *cobra.Command,
	args []string,
) error {
	return func(_ *cobra.Command, _ []string) error {
		// create home directory
		home, err := rootCmd.PersistentFlags().GetString(flags.FlagHome)
		if err != nil {
			return err
		}
		if err = os.MkdirAll(home, os.ModePerm); err != nil {
			return err
		}

		// init temporary application
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

		// set keyring
		keyring, err := keyring.New("band", keyring.BackendTest, home, nil, tempApplication.AppCodec())
		if err != nil {
			return err
		}

		newCtx, err := context.NewContext(
			nil,
			keyring,
			home,
			tempApplication.AppCodec(),
			tempApplication.GetTxConfig(),
			tempApplication.InterfaceRegistry(),
		)
		if err != nil {
			return err
		}
		*ctx = *newCtx

		return initConfig(ctx, rootCmd)
	}
}

func getDefaultHome() string {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(userHomeDir, ".cylinder")
}

var tempDir = func() string {
	dir, err := os.MkdirTemp("", ".band")
	if err != nil {
		dir = band.DefaultNodeHome
	}
	defer os.RemoveAll(dir)

	return dir
}
