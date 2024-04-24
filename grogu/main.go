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
	"github.com/bandprotocol/chain/v2/grogu/cmd"
	grogucontext "github.com/bandprotocol/chain/v2/grogu/context"
)

func initConfig(c *grogucontext.Context, cmd *cobra.Command) error {
	viper.SetConfigFile(path.Join(c.Home, "config.yaml"))
	_ = viper.ReadInConfig() // If we fail to read config file, we'll just rely on cmd flags.
	if err := viper.Unmarshal(&grogucontext.Cfg); err != nil {
		return err
	}
	return nil
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	grogucontext.DefaultGroguHome = filepath.Join(userHomeDir, ".grogu")
}

func Main() {
	appConfig := sdk.GetConfig()
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(appConfig)

	ctx := &grogucontext.Context{}
	rootCmd := &cobra.Command{
		Use:   "grogu",
		Short: "BandChain daemon to submit prices for feeds module",
	}

	rootCmd.AddCommand(
		cmd.ConfigCmd(),
		cmd.KeysCmd(ctx),
		cmd.RunCmd(ctx),
		version.NewVersionCommand(),
	)
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		home, err := rootCmd.PersistentFlags().GetString(flags.FlagHome)
		if err != nil {
			return err
		}
		ctx.Home = home
		if err := os.MkdirAll(home, os.ModePerm); err != nil {
			return err
		}
		grogucontext.Kb, err = keyring.New("band", keyring.BackendTest, home, nil, grogucontext.Cdc)
		if err != nil {
			return err
		}
		return initConfig(ctx, rootCmd)
	}
	rootCmd.PersistentFlags().String(flags.FlagHome, grogucontext.DefaultGroguHome, "home directory")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
