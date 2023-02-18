package yoda

import (
	"bufio"
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	bip39 "github.com/cosmos/go-bip39"
	"github.com/kyokomi/emoji"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	band "github.com/bandprotocol/chain/v2/app"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

const (
	flagAccount  = "account"
	flagIndex    = "index"
	flagCoinType = "coin-type"
	flagRecover  = "recover"
	flagAddress  = "address"
)

func keysCmd(c *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "keys",
		Aliases: []string{"k"},
		Short:   "Manage key held by the oracle process",
	}
	cmd.AddCommand(
		keysAddCmd(c),
		keysDeleteCmd(c),
		keysListCmd(c),
		keysShowCmd(c),
	)
	return cmd
}

func keysAddCmd(c *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add [name]",
		Aliases: []string{"a"},
		Short:   "Add a new key to the keychain",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var mnemonic string
			recover, err := cmd.Flags().GetBool(flagRecover)
			if err != nil {
				return err
			}
			if recover {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				var err error
				mnemonic, err = input.GetString("Enter your bip39 mnemonic", inBuf)
				if err != nil {
					return err
				}
			} else {
				seed, err := bip39.NewEntropy(256)
				if err != nil {
					return err
				}
				mnemonic, err = bip39.NewMnemonic(seed)
				if err != nil {
					return err
				}
				fmt.Printf("Mnemonic: %s\n", mnemonic)
			}

			if err != nil {
				return err
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
			info, err := kb.NewAccount(args[0], mnemonic, "", hdPath.String(), hd.Secp256k1)
			if err != nil {
				return err
			}

			address, err := info.GetAddress()
			if err != nil {
				return err
			}

			fmt.Printf("Address: %s\n", address.String())
			return nil
		},
	}
	cmd.Flags().Bool(flagRecover, false, "Provide seed phrase to recover existing key instead of creating")
	cmd.Flags().Uint32(flagAccount, 0, "Account number for HD derivation")
	cmd.Flags().Uint32(flagIndex, 0, "Address index number for HD derivation")
	return cmd
}

func keysDeleteCmd(c *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [name]",
		Aliases: []string{"d"},
		Short:   "Delete a key from the keychain",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			_, err := kb.Key(name)
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

			if err := kb.Delete(name); err != nil {
				return err
			}

			fmt.Printf("Deleted key: %s\n", name)
			return nil
		},
	}
	return cmd
}

func keysListCmd(c *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "List all the keys in the keychain",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			keys, err := kb.List()
			if err != nil {
				return err
			}
			isShowAddr := viper.GetBool(flagAddress)
			for _, key := range keys {
				address, err := key.GetAddress()
				if err != nil {
					return err
				}

				if isShowAddr {
					fmt.Printf("%s ", address.String())
				} else {
					queryClient := types.NewQueryClient(clientCtx)
					r, err := queryClient.IsReporter(
						context.Background(),
						&types.QueryIsReporterRequest{ValidatorAddress: cfg.Validator, ReporterAddress: address.String()},
					)
					s := ":question:"
					if err == nil {
						if r.IsReporter {
							s = ":white_check_mark:"
						} else {
							s = ":x:"
						}
					}
					emoji.Printf("%s%s => %s\n", s, key.Name, address.String())
				}
			}

			return nil
		},
	}
	cmd.Flags().BoolP(flagAddress, "a", false, "Output the address only")
	viper.BindPFlag(flagAddress, cmd.Flags().Lookup(flagAddress))

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func keysShowCmd(c *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show [name]",
		Aliases: []string{"s"},
		Short:   "Show address from name in the keychain",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			key, err := kb.Key(name)
			if err != nil {
				return err
			}

			address, err := key.GetAddress()
			if err != nil {
				return err
			}
			fmt.Println(address.String())
			return nil
		},
	}
	return cmd
}
