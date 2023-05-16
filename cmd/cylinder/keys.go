package main

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

	band "github.com/bandprotocol/chain/v2/app"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

const (
	flagAccount  = "account"
	flagIndex    = "index"
	flagCoinType = "coin-type"
	flagRecover  = "recover"
	flagAddress  = "address"
)

// keysCmd returns a Cobra command for managing keys.
func keysCmd(ctx *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "keys",
		Aliases: []string{"k"},
		Short:   "Manage key held by the cylinder process",
	}

	cmd.AddCommand(
		keysAddCmd(ctx),
		keysDeleteCmd(ctx),
		keysListCmd(ctx),
		keysShowCmd(ctx),
	)

	return cmd
}

// keysAddCmd returns a Cobra command for adding a new key to the keychain.
func keysAddCmd(ctx *Context) *cobra.Command {
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

			account, err := cmd.Flags().GetUint32(flagAccount)
			if err != nil {
				return err
			}

			index, err := cmd.Flags().GetUint32(flagIndex)
			if err != nil {
				return err
			}

			hdPath := hd.CreateHDPath(band.Bip44CoinType, account, index)
			info, err := ctx.keyring.NewAccount(args[0], mnemonic, "", hdPath.String(), hd.Secp256k1)
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

// keysDeleteCmd returns a Cobra command for deleting a key from the keychain.
func keysDeleteCmd(ctx *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [name]",
		Aliases: []string{"d"},
		Short:   "Delete a key from the keychain",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Retrieve the key from the keyring
			_, err := ctx.keyring.Key(name)
			if err != nil {
				return err
			}

			inBuf := bufio.NewReader(cmd.InOrStdin())

			// Ask for confirmation from the user
			confirmInput, err := input.GetString("Key will be deleted. Continue?[y/N]", inBuf)
			if err != nil {
				return err
			}

			if confirmInput != "y" {
				fmt.Println("Cancel")
				return nil
			}

			// Delete the key from the keyring
			if err := ctx.keyring.Delete(name); err != nil {
				return err
			}

			fmt.Printf("Deleted key: %s\n", name)
			return nil
		},
	}

	return cmd
}

// keysListCmd returns a Cobra command for listing all the keys in the keychain.
func keysListCmd(ctx *Context) *cobra.Command {
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

			// Retrieve the list of keys from the keyring
			keys, err := ctx.keyring.List()
			if err != nil {
				return err
			}

			// Check if the "--address" flag is provided
			isShowAddr, err := cmd.Flags().GetBool(flagAddress)
			if err != nil {
				return err
			}

			for _, key := range keys {
				// Retrieve the address associated with the key
				address, err := key.GetAddress()
				if err != nil {
					return err
				}

				if isShowAddr {
					fmt.Printf("%s ", address.String())
				} else {
					// Query if the key is a grantee and display the result
					queryClient := types.NewQueryClient(clientCtx)
					r, err := queryClient.IsGrantee(
						context.Background(),
						&types.QueryIsGranteeRequest{GranterAddress: ctx.config.Granter, GranteeAddress: address.String()},
					)
					s := ":question:"
					if err == nil {
						if r.IsGrantee {
							s = ":white_check_mark:"
						} else {
							s = ":x:"
						}
						emoji.Printf("%s%s => %s\n", s, key.Name, address.String())
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolP(flagAddress, "a", false, "Output the address only")

	// Add the query flags to the command
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// keysShowCmd returns a Cobra command for showing the address associated with a key in the keychain.
func keysShowCmd(ctx *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show [name]",
		Aliases: []string{"s"},
		Short:   "Show address from name in the keychain",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Retrieve the key from the keyring
			key, err := ctx.keyring.Key(name)
			if err != nil {
				return err
			}

			// Retrieve the address associated with the key
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
