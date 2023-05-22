package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// GetQueryCmd returns the cli query commands for the tss module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the tss module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(QueryIsGrantee())
	cmd.AddCommand(QueryGroupCmd())
	cmd.AddCommand(QueryMembersCmd())

	return cmd
}

// QueryIsGrantee creates a CLI command for Query/IsGrantee.
func QueryIsGrantee() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "is-grantee [granter_address] [grantee_address]",
		Short: "Query grantee status",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.IsGrantee(cmd.Context(), &types.QueryIsGranteeRequest{
				GranterAddress: args[0],
				GranteeAddress: args[1],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryGroupCmd creates a CLI command for Query/Group.
func QueryGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group [id]",
		Short: "Query group by group id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Group(cmd.Context(), &types.QueryGroupRequest{
				GroupId: groupID,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryMembersCmd creates a CLI command for Query/Members.
func QueryMembersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "members [group-id]",
		Short: "Query members by group id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Members(cmd.Context(), &types.QueryMembersRequest{
				GroupId: groupID,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
