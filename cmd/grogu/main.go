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
	"github.com/bandprotocol/chain/v2/cmd/grogu/cmd"
	"github.com/bandprotocol/chain/v2/grogu/context"
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
		ctx.Home = home

		if err = os.MkdirAll(home, os.ModePerm); err != nil {
			return err
		}

		cdc := band.MakeEncodingConfig().Marshaler
		ctx.Keyring, err = keyring.New("band", keyring.BackendTest, home, nil, cdc)
		if err != nil {
			return err
		}

		return initConfig(ctx, rootCmd)
	}
}

func initConfig(c *context.Context, _ *cobra.Command) error {
	viper.SetConfigFile(path.Join(c.Home, "config.yaml"))

	// If the config file cannot be read, only cmd flags will be used.
	_ = viper.ReadInConfig()
	if err := viper.Unmarshal(&c.Config); err != nil {
		return err
	}
	return nil
}

func getDefaultHome() string {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(userHomeDir, ".grogu")
}
