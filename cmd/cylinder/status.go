package main

import (
	"context"
	"fmt"

	"github.com/kyokomi/emoji"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	cylinderclient "github.com/bandprotocol/chain/v3/cylinder/client"
	cylinderctx "github.com/bandprotocol/chain/v3/cylinder/context"
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
)

// statusCmd returns a cobra command to show the tss member status of the given address.
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
			member, err := queryClient.Member(context.Background(), &types.QueryMemberRequest{
				Address: address,
			})
			if err != nil {
				return err
			}

			memberResponse := cylinderclient.NewMemberResponse(member)
			if len(memberResponse.Members) == 0 {
				return fmt.Errorf("no members found for address %s", address)
			}

			for _, member := range memberResponse.Members {
				status := ":white_check_mark:"
				if !member.IsActive {
					status = ":x:"
				}
				emoji.Printf("group %d with member %s is active: %s", member.GroupID, address, status)
			}

			return nil
		},
	}

	// Add the query flags to the command
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
