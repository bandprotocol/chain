package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

const (
	flagMemberStatusFilter = "status"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the bandtss module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetQueryCmdCounts(),
		GetQueryCmdMember(),
		GetQueryCmdMembers(),
		GetQueryCmdCurrentGroup(),
		GetQueryCmdIncomingGroup(),
		GetQueryCmdParams(),
		GetQueryCmdSigning(),
		GetQueryCmdGroupTransition(),
		GetQueryCmdIsGrantee(),
	)

	return cmd
}

// GetQueryCmdCounts implements the query counts command.
func GetQueryCmdCounts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "counts",
		Short: "Get current number of signing requests to bandtss module on BandChain",
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

// GetQueryCmdMember creates a CLI command for querying member information.
func GetQueryCmdMember() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member [address]",
		Short: "Query the member by address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Member(cmd.Context(), &types.QueryMemberRequest{
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

// GetQueryCmdMembers creates a CLI command for querying members information.
func GetQueryCmdMembers() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "members",
		Short: "Query the members information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			statusFilterFlag, err := cmd.Flags().GetBool(flagMemberStatusFilter)
			if err != nil {
				return err
			}

			isIncomingGroup, err := cmd.Flags().GetBool(flagIncomingGroup)
			if err != nil {
				return err
			}

			var statusFilter types.MemberStatusFilter
			if !cmd.Flags().Changed(flagMemberStatusFilter) {
				statusFilter = types.MEMBER_STATUS_FILTER_UNSPECIFIED
			} else if statusFilterFlag {
				statusFilter = types.MEMBER_STATUS_FILTER_ACTIVE
			} else {
				statusFilter = types.MEMBER_STATUS_FILTER_INACTIVE
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Members(cmd.Context(), &types.QueryMembersRequest{
				Status:          statusFilter,
				IsIncomingGroup: isIncomingGroup,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().
		Bool(flagIncomingGroup, false, "Whether the heartbeat is for the incoming group or current group.")
	cmd.Flags().Bool(flagMemberStatusFilter, false, "Filter members by status")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryCmdCurrentGroup creates a CLI command for querying current group.
func GetQueryCmdCurrentGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current-group",
		Short: "Query the current group information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.CurrentGroup(cmd.Context(), &types.QueryCurrentGroupRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryCmdIncomingGroup creates a CLI command for querying incoming group.
func GetQueryCmdIncomingGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "incoming-group",
		Short: "Query the incoming group information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.IncomingGroup(cmd.Context(), &types.QueryIncomingGroupRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryCmdParams creates a CLI command for querying module's parameter.
func GetQueryCmdParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Show params",
		Long:  "Show parameter of bandtss module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetQueryCmdSigning creates a CLI command for querying signing information.
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

// GetQueryCmdGroupTransition creates a CLI command for querying group transition information.
func GetQueryCmdGroupTransition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group-transition",
		Short: "Query the group transition information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.GroupTransition(cmd.Context(), &types.QueryGroupTransitionRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryCmdIsGrantee creates a CLI command for querying whether a grantee is granted by a granter.
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
