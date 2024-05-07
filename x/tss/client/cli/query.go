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

	cmd.AddCommand(
		GetQueryCmdCounts(),
		GetQueryCmdGroup(),
		GetQueryCmdMembers(),
		GetQueryCmdIsGrantee(),
		GetQueryCmdDE(),
		GetQueryCmdPendingSignings(),
		GetQueryCmdSigning(),
	)

	return cmd
}

// GetQueryCmdCounts implements the query counts command.
func GetQueryCmdCounts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "counts",
		Short: "Get current number of groups and signings on BandChain",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Counts(cmd.Context(), &types.QueryCountsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryCmdGroup creates a CLI command for Query/Group.
func GetQueryCmdGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group [id]",
		Short: "Query group by group ID",
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

// GetQueryCmdMembers creates a CLI command for Query/Members.
func GetQueryCmdMembers() *cobra.Command {
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

// GetQueryCmdIsGrantee creates a CLI command for Query/IsGrantee.
func GetQueryCmdIsGrantee() *cobra.Command {
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
				Granter: args[0],
				Grantee: args[1],
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

// GetQueryCmdDE creates a CLI command for Query/DE.
func GetQueryCmdDE() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "de-list [address]",
		Short: "Query all DE for this address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.DE(cmd.Context(), &types.QueryDERequest{
				Address: args[0],
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

// GetQueryCmdPendingSignings creates a CLI command for Query/PendingSignings.
func GetQueryCmdPendingSignings() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-signings [address]",
		Short: "Query all pending signing for this address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.PendingSignings(cmd.Context(), &types.QueryPendingSigningsRequest{
				Address: args[0],
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

// GetQueryCmdSigning creates a CLI command for Query/Signing.
func GetQueryCmdSigning() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signing [id]",
		Short: "Query a signing by signing ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			signingID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Signing(cmd.Context(), &types.QuerySigningRequest{
				SigningId: signingID,
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
