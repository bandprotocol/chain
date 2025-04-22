package main

import (
	"context"

	"github.com/kyokomi/emoji"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	cylinderctx "github.com/bandprotocol/chain/v3/cylinder/context"
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
)

// statusCmd returns
func statusCmd(ctx *cylinderctx.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status [address]",
		Aliases: []string{"s"},
		Short:   "Show the tss member status of the given address",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			address := args[0]

			queryClient := types.NewQueryClient(clientCtx)
			currentGroup, err := queryClient.Members(context.Background(), &types.QueryMembersRequest{})
			if err != nil {
				return err
			}
			incomingGroup, err := queryClient.Members(
				context.Background(),
				&types.QueryMembersRequest{IsIncomingGroup: true},
			)
			if err != nil {
				return err
			}

			members := append(currentGroup.Members, incomingGroup.Members...)
			for _, member := range members {
				if member.Address == address {
					status := ":white_check_mark:"
					if !member.IsActive {
						status = ":x:"
					}
					emoji.Printf("group %d with member %s is active: %s", member.GroupID, address, status)
				}
			}
			return nil
		},
	}

	// Add the query flags to the command
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
