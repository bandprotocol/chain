package main

import (
	"encoding/hex"
	"os"
	"path"
	"path/filepath"
	"reflect"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	band "github.com/bandprotocol/chain/v2/app"
	"github.com/bandprotocol/chain/v2/pkg/tss"
)

// Global instances.
var (
	DefaultHome = filepath.Join(os.Getenv("HOME"), ".cylinder")
	cdc         = band.MakeEncodingConfig().Marshaler
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
func initConfig(ctx *Context, cmd *cobra.Command) error {
	var err error
	if err := os.MkdirAll(ctx.home, os.ModePerm); err != nil {
		return err
	}

	ctx.keyring, err = keyring.New("band", keyring.BackendTest, ctx.home, nil, cdc)
	if err != nil {
		return err
	}

	viper.SetConfigFile(path.Join(ctx.home, "config.yaml"))
	_ = viper.ReadInConfig() // If we fail to read config file, we'll just rely on cmd flags.

	configOption := viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		hexByteToScalarHookFunc(),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	))

	if err := viper.Unmarshal(ctx.config, configOption); err != nil {
		return err
	}

	return nil
}

// NewRootCmd returns a new instance of the root command.
func NewRootCmd() *cobra.Command {
	ctx := &Context{}
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

	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(ctx, rootCmd)
	}

	rootCmd.PersistentFlags().StringVar(&ctx.home, flags.FlagHome, DefaultHome, "home directory")

	return rootCmd
}
