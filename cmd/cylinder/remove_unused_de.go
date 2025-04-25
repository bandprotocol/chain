package main

import (
	"encoding/base64"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/bandprotocol/chain/v3/cylinder/client"
	"github.com/bandprotocol/chain/v3/cylinder/context"
)

// removeUnusedDECmd returns a Cobra command for removing unused DEs from the store
func removeUnusedDECmd(ctx *context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-unused-de",
		Short: "Remove DEs from the store that are not present in the chain",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize the database
			ctx, err := ctx.WithGoLevelDB()
			if err != nil {
				return err
			}

			// Initialize the client
			cli, err := client.New(ctx)
			if err != nil {
				return err
			}

			// Get all DEs from the store
			storeDEs, err := ctx.Store.GetAllDEs()
			if err != nil {
				return err
			}

			// Get DE information from the chain
			chainDEs, err := cli.QueryAllDE(ctx.Config.Granter)
			if err != nil {
				return err
			}

			chainDEsMap := make(map[string]bool)
			for _, de := range chainDEs {
				chainDEsMap[de.String()] = true
			}

			// Remove DEs that are not in the chain
			removedCount := 0
			for _, de := range storeDEs {
				if !chainDEsMap[de.PubDE.String()] {
					if err := ctx.Store.DeleteDE(de.PubDE); err != nil {
						return err
					}
					removedCount++

					pubD := base64.StdEncoding.EncodeToString(de.PubDE.PubD.Bytes())
					pubE := base64.StdEncoding.EncodeToString(de.PubDE.PubE.Bytes())
					ctx.Logger.Info(":white_check_mark: Removed DE from the store (PubD: %s PubE: %s)", pubD, pubE)
				}
			}

			ctx.Logger.Info(":white_check_mark: Successfully removed %d unused DEs from the store", removedCount)
			return nil
		},
	}

	// Add the query flags to the command
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
