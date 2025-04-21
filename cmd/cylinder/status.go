package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/kyokomi/emoji"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	cylinderclient "github.com/bandprotocol/chain/v3/cylinder/client"
	cylinderctx "github.com/bandprotocol/chain/v3/cylinder/context"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// statusCmd returns a Cobra command for showing the tss status of the given group id and address.
func statusCmd(ctx *cylinderctx.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status [group-id] [address]",
		Aliases: []string{"s"},
		Short:   "Show the tss member status of the given group id and address",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Print the node URI for debugging
			fmt.Printf("Node URI: %s\n", ctx.Config.NodeURI)

			// Set the node URI in the command flags
			if err := cmd.Flags().Set(flags.FlagNode, ctx.Config.NodeURI); err != nil {
				return fmt.Errorf("failed to set node URI: %w", err)
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			address := args[1]

			queryClient := types.NewQueryClient(clientCtx)
			r, err := queryClient.Members(
				context.Background(),
				&types.QueryMembersRequest{GroupId: groupID},
			)
			if err != nil {
				return err
			}

			members := cylinderclient.NewMembersResponse(r)
			isActive, err := members.IsActive(address)
			if err != nil {
				return err
			}

			status := ":white_check_mark:"
			if !isActive {
				status = ":x:"
			}
			emoji.Printf("group %d with member %s is active: %s\n", groupID, address, status)

			return nil
		},
	}

	// Add the query flags to the command
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
