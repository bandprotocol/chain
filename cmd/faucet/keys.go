package main

import (
	"bufio"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/go-bip39"
	"github.com/spf13/cobra"

	band "github.com/bandprotocol/chain/v2/app"
)

const (
	flagAccount  = "account"
	flagIndex    = "index"
	flagCoinType = "coin-type"
	flagRecover  = "recover"
)

func keysCmd(c *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "keys",
		Aliases: []string{"k"},
		Short:   "Manage key held by the oracle process",
	}
	cmd.AddCommand(keysAddCmd(c))
	cmd.AddCommand(keysDeleteCmd(c))
	cmd.AddCommand(keysListCmd(c))
	cmd.AddCommand(keysShowCmd(c))
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
			info, err := keybase.NewAccount(args[0], mnemonic, "", hdPath.String(), hd.Secp256k1)
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
			_, err := keybase.Key(name)
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
			if err := keybase.Delete(name); err != nil {
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
			keys, err := keybase.List()
			if err != nil {
				return err
			}
			for _, key := range keys {
				address, err := key.GetAddress()
				if err != nil {
					return err
				}

				fmt.Printf("%s => %s\n", key.Name, address.String())
			}
			return nil
		},
	}
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

			key, err := keybase.Key(name)
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
