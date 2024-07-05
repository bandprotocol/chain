package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	"github.com/kyokomi/emoji"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	band "github.com/bandprotocol/chain/v2/app"
	grogu "github.com/bandprotocol/chain/v2/grogu/context"
	"github.com/bandprotocol/chain/v2/grogu/querier"
)

const (
	flagAccount = "account"
	flagIndex   = "index"
	flagRecover = "recover"
	flagAddress = "address"
)

func KeysCmd(ctx *grogu.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "keys",
		Aliases: []string{"k"},
		Short:   "Manage keys held by Grogu",
	}
	cmd.AddCommand(
		keysAddCmd(ctx),
		keysDeleteCmd(ctx),
		keysListCmd(ctx),
		keysShowCmd(ctx),
	)
	return cmd
}

func keysAddCmd(ctx *grogu.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add [name]",
		Aliases: []string{"a"},
		Short:   "Add a new key to the keychain",
		Args:    cobra.ExactArgs(1),
		RunE:    createKeysAddRunE(ctx),
	}

	cmd.Flags().Bool(flagRecover, false, "Provide seed phrase to recover existing key instead of creating")
	cmd.Flags().Uint32(flagAccount, 0, "Account number for HD derivation")
	cmd.Flags().Uint32(flagIndex, 0, "Address index number for HD derivation")

	return cmd
}

func createKeysAddRunE(ctx *grogu.Context) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var mnemonic string
		var passphrase string
		var err error

		recoverKey, err := cmd.Flags().GetBool(flagRecover)
		if err != nil {
			return err
		}

		if recoverKey {
			inBuf := bufio.NewReader(cmd.InOrStdin())

			mnemonic, err = input.GetString("Enter your bip39 mnemonic", inBuf)
			if err != nil {
				return err
			}

			passphrase, err = input.GetString("Enter your bip39 passphrase", inBuf)
			if err != nil {
				return err
			}
		} else {
			var seed []byte
			seed, err = bip39.NewEntropy(256)
			if err != nil {
				return err
			}

			mnemonic, err = bip39.NewMnemonic(seed)
			if err != nil {
				return err
			}
			fmt.Printf("Mnemonic: %s\n", mnemonic)
		}

		account, err := cmd.Flags().GetUint32(flagAccount)
		if err != nil {
			return err
		}

		index, err := cmd.Flags().GetUint32(flagIndex)
		if err != nil {
			return err
		}

		hdPath := hd.CreateHDPath(band.Bip44CoinType, account, index)
		info, err := ctx.Keyring.NewAccount(args[0], mnemonic, passphrase, hdPath.String(), hd.Secp256k1)
		if err != nil {
			return err
		}

		address, err := info.GetAddress()
		if err != nil {
			return err
		}

		fmt.Printf("Address: %s\n", address.String())
		return nil
	}
}

func keysDeleteCmd(ctx *grogu.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [name]",
		Aliases: []string{"d"},
		Short:   "Delete a key from the keychain",
		Args:    cobra.ExactArgs(1),
		RunE:    createKeysDeleteRunE(ctx),
	}

	return cmd
}

func createKeysDeleteRunE(ctx *grogu.Context) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		name := args[0]

		_, err := ctx.Keyring.Key(name)
		if err != nil {
			return err
		}

		inBuf := bufio.NewReader(cmd.InOrStdin())
		confirmInput, err := input.GetString("Key will be deleted. Continue?[y/N]", inBuf)
		if err != nil {
			return err
		}

		if confirmInput != "y" {
			fmt.Println("Cancel")
			return nil
		}

		if err = ctx.Keyring.Delete(name); err != nil {
			return err
		}

		fmt.Printf("Deleted key: %s\n", name)
		return nil
	}
}

func keysListCmd(ctx *grogu.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "List all the keys in the keychain",
		Args:    cobra.ExactArgs(0),
		RunE:    createKeysListRunE(ctx),
	}

	cmd.Flags().BoolP(flagAddress, "a", false, "Output the address only")
	_ = viper.BindPFlag(flagAddress, cmd.Flags().Lookup(flagAddress))

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func createKeysListRunE(ctx *grogu.Context) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		bandConfig := band.MakeEncodingConfig()

		nodesURIs := strings.Split(viper.GetString(flagNodes), ",")
		clients := make([]client.Context, 0, len(nodesURIs))
		for _, URI := range nodesURIs {
			httpClient, err := http.New(URI, "/websocket")
			if err != nil {
				continue
			}
			cl := client.Context{
				Client:            httpClient,
				ChainID:           ctx.Config.ChainID,
				Codec:             bandConfig.Marshaler,
				InterfaceRegistry: bandConfig.InterfaceRegistry,
				Keyring:           ctx.Keyring,
				TxConfig:          bandConfig.TxConfig,
				BroadcastMode:     flags.BroadcastSync,
			}
			clients = append(clients, cl)
		}

		feedQuerier := querier.NewFeedQuerier(clients)

		keys, err := ctx.Keyring.List()
		if err != nil {
			return err
		}
		isShowAddr := viper.GetBool(flagAddress)

		validatorAddr, err := sdk.ValAddressFromBech32(ctx.Config.Validator)
		if err != nil {
			fmt.Printf("Invalid validator address: %s\n", ctx.Config.Validator)
		}

		for _, key := range keys {
			address, err := key.GetAddress()
			if err != nil {
				return err
			}

			if isShowAddr {
				fmt.Printf("%s ", address.String())
				continue
			}

			resp, err := feedQuerier.QueryIsFeeder(validatorAddr, address)

			s := ":question:"
			if err == nil {
				if resp.IsFeeder {
					s = ":white_check_mark:"
				} else {
					s = ":x:"
				}
			}
			_, err = emoji.Printf("%s%s => %s\n", s, key.Name, address.String())
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func keysShowCmd(ctx *grogu.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show [name]",
		Aliases: []string{"s"},
		Short:   "Show address from name in the keychain",
		Args:    cobra.ExactArgs(1),
		RunE:    createKeysShowRunE(ctx),
	}

	return cmd
}

func createKeysShowRunE(ctx *grogu.Context) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		name := args[0]

		key, err := ctx.Keyring.Key(name)
		if err != nil {
			return err
		}

		address, err := key.GetAddress()
		if err != nil {
			return err
		}

		fmt.Println(address.String())
		return nil
	}
}
