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
		Use:     "status",
		Aliases: []string{"s"},
		Short:   "Show the TSS member status of the given address",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			address := ctx.Config.Granter
			if address == "" {
				return fmt.Errorf("granter address is not set in the config")
			}

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
				if !member.IsActive {
					emoji.Printf(":warning:group %d with member %s is inactive", member.GroupID, address)
				} else {
					emoji.Printf(":white_check_mark:group %d with member %s is active", member.GroupID, address)
				}
			}

			return nil
		},
	}

	// Add the query flags to the command
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
