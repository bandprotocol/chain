package yoda

import (
	"bufio"
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/go-bip39"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	odin "github.com/GeoDB-Limited/odin-core/app"
)

const (
	flagAccount = "account"
	flagIndex   = "index"
	flagRecover = "recover"
	flagAddress = "address"
)

const (
	EntropyBitSize = 256
)

func keysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "keys",
		Aliases: []string{"k"},
		Short:   "Manage key held by the oracle process",
	}
	cmd.AddCommand(
		keysAddCmd(),
		keysDeleteCmd(),
		keysListCmd(),
		keysShowCmd(),
	)
	return cmd
}

func keysAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add [name]",
		Aliases: []string{"a"},
		Short:   "Add a new key to the keychain",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var mnemonic string
			rec, err := cmd.Flags().GetBool(flagRecover)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to parse rec flag")
			}
			if rec {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				mnemonic, err = input.GetString("Enter your bip39 mnemonic", inBuf)
				if err != nil {
					return sdkerrors.Wrapf(err, "failed to parse input mnemonic")
				}
			} else {
				seed, err := bip39.NewEntropy(EntropyBitSize)
				if err != nil {
					return sdkerrors.Wrapf(err, "failed to create a new entropy")
				}
				mnemonic, err = bip39.NewMnemonic(seed)
				if err != nil {
					return sdkerrors.Wrapf(err, "failed to create a new mnemonic with the given seed")
				}
				fmt.Printf("Mnemonic: %s\n", mnemonic)
			}

			account, err := cmd.Flags().GetUint32(flagAccount)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to parse account flag")
			}
			index, err := cmd.Flags().GetUint32(flagIndex)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to parse index flag")
			}
			hdPath := hd.CreateHDPath(odin.Bip44CoinType, account, index)
			info, err := yoda.keybase.NewAccount(args[0], mnemonic, "", hdPath.String(), hd.Secp256k1)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to create a new keyring account")
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

func keysDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [name]",
		Aliases: []string{"d"},
		Short:   "Delete a key from the keychain",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			_, err := yoda.keybase.Key(name)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to find value by the given key: %s", name)
			}

			inBuf := bufio.NewReader(cmd.InOrStdin())
			confirmInput, err := input.GetString("Key will be deleted. Continue?[y/N]", inBuf)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to parse confirmation input")
			}

			if confirmInput != "y" {
				fmt.Println("Cancel")
				return nil
			}

			if err := yoda.keybase.Delete(name); err != nil {
				return sdkerrors.Wrapf(err, "failed to delete value by the given key: %s", name)
			}

			fmt.Printf("Deleted key: %s\n", name)
			return nil
		},
	}
	return cmd
}

func keysListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "List all the keys in the keychain",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			keys, err := yoda.keybase.List()
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to get keys from keyring")
			}
			isShowAddr := viper.GetBool(flagAddress)
			for _, key := range keys {
				if isShowAddr {
					fmt.Printf("%s ", key.GetAddress().String())
				} else {
					fmt.Printf("%s => %s\n", key.GetName(), key.GetAddress().String())
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolP(flagAddress, "a", false, "Output the address only")
	if err := viper.BindPFlag(flagAddress, cmd.Flags().Lookup(flagAddress)); err != nil {
		panic(sdkerrors.Wrapf(err, "failed to parse %s flag", flagAddress))
	}

	return cmd
}

func keysShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show [name]",
		Aliases: []string{"s"},
		Short:   "Show address from name in the keychain",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			key, err := yoda.keybase.Key(name)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to get value by the given key: %s", name)
			}
			fmt.Println(key.GetAddress().String())
			return nil
		},
	}
	return cmd
}
