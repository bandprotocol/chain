package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

// GetQueryCmd returns the cli query commands for the bandtss module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the bandtss module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetQueryCmdMember(),
		GetQueryCmdCurrentGroup(),
		GetQueryCmdReplacingGroup(),
		GetQueryCmdParams(),
		GetQueryCmdSigning(),
	)

	return cmd
}

// GetQueryCmdMember creates a CLI command for Query/Member.
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

// GetQueryCmdCurrentGroup creates a CLI command for querying current group.
func GetQueryCmdCurrentGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current-group",
		Short: "Query the currentGroup",
		Args:  cobra.ExactArgs(0),
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

// GetQueryCmdReplacingGroup creates a CLI command for querying replacing group.
func GetQueryCmdReplacingGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replacing-group",
		Short: "Query the replacingGroup",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.ReplacingGroup(cmd.Context(), &types.QueryReplacingGroupRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetQueryCmdParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Show params",
		Long:  "Show parameter of bandtss module",
		Args:  cobra.ExactArgs(0),
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
				SigningID: types.SigningID(signingID),
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
