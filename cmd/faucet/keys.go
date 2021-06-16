package main

import (
	"bufio"
	"fmt"
	band "github.com/GeoDB-Limited/odin-core/app"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/go-bip39"
	"github.com/spf13/cobra"
)

const (
	flagAccount = "account"
	flagIndex   = "index"
	flagRecover = "recover"
)

const (
	EntropySize = 256
)

// KeysCmd defines the list of commands to manage faucet keybase.
func KeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "keys",
		Aliases: []string{"k"},
		Short:   "Manage key held by the oracle process",
	}
	cmd.AddCommand(faucet.keysAddCmd())
	cmd.AddCommand(faucet.keysDeleteCmd())
	cmd.AddCommand(faucet.keysListCmd())
	cmd.AddCommand(faucet.keysShowCmd())
	return cmd
}

// keysAddCmd adds new key to the keybase.
func (f *Faucet) keysAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add [name]",
		Aliases: []string{"a"},
		Short:   "Add a new key to the keychain",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			account, err := cmd.Flags().GetUint32(flagAccount)
			if err != nil {
				return sdkerrors.Wrap(err, "failed to parse the account flag")
			}
			index, err := cmd.Flags().GetUint32(flagIndex)
			if err != nil {
				return sdkerrors.Wrap(err, "failed to parse the index flag")
			}
			rec, err := cmd.Flags().GetBool(flagRecover)
			if err != nil {
				return sdkerrors.Wrap(err, "failed to parse recover flag")
			}

			var mnemonic string
			if rec {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				mnemonic, err = input.GetString("Enter your bip39 mnemonic", inBuf)
				if err != nil {
					return sdkerrors.Wrap(err, "failed to parse the mnemonic")
				}
			} else {
				seed, err := bip39.NewEntropy(EntropySize)
				if err != nil {
					return sdkerrors.Wrap(err, "failed to create a new entropy")
				}
				mnemonic, err = bip39.NewMnemonic(seed)
				if err != nil {
					return sdkerrors.Wrap(err, "failed to create a new mnemonic")
				}
				fmt.Printf("Mnemonic: %s\n", mnemonic)
			}

			hdPath := hd.CreateHDPath(band.Bip44CoinType, account, index)
			info, err := f.keybase.NewAccount(args[0], mnemonic, "", hdPath.String(), hd.Secp256k1)
			if err != nil {
				return sdkerrors.Wrap(err, "failed to create a new keybase account")
			}
			fmt.Printf("Address: %s\n", info.GetAddress().String())
			return nil
		},
	}

	cmd.Flags().Bool(flagRecover, false, "Provide seed phrase to recover existing key instead of creating")
	cmd.Flags().Uint32(flagAccount, 0, "Account number for HD derivation")
	cmd.Flags().Uint32(flagIndex, 0, "Address index number for HD derivation")

	return cmd
}

// keysDeleteCmd removes key form the keybase by the given key.
func (f *Faucet) keysDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [name]",
		Aliases: []string{"d"},
		Short:   "Delete a key from the keychain",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			_, err := f.keybase.Key(name)
			if err != nil {
				return sdkerrors.Wrap(err, "failed to get by the given key")
			}

			inBuf := bufio.NewReader(cmd.InOrStdin())
			confirmInput, err := input.GetString("Key will be deleted. Continue?[y/N]", inBuf)
			if err != nil {
				return sdkerrors.Wrap(err, "failed to read the input")
			}

			if confirmInput != "y" {
				fmt.Println("Cancel")
				return nil
			}

			if err := f.keybase.Delete(name); err != nil {
				return sdkerrors.Wrap(err, "failed to delete by the given key")
			}

			fmt.Printf("Deleted key: %s\n", name)
			return nil
		},
	}
	return cmd
}

// keysListCmd prints all values from the keybase.
func (f *Faucet) keysListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "List all the keys in the keychain",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			keys, err := f.keybase.List()
			if err != nil {
				return sdkerrors.Wrap(err, "failed to retrieve the keys list")
			}
			for _, key := range keys {
				fmt.Printf("%s => %s\n", key.GetName(), key.GetAddress().String())
			}
			return nil
		},
	}
	return cmd
}

// keysShowCmd prints value by the given key.
func (f *Faucet) keysShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show [name]",
		Aliases: []string{"s"},
		Short:   "Show address from name in the keychain",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			key, err := f.keybase.Key(name)
			if err != nil {
				return sdkerrors.Wrap(err, "failed to get by the given key")
			}
			fmt.Println(key.GetAddress().String())
			return nil
		},
	}
	return cmd
}
