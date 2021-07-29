package cli

import (
	telemetrytypes "github.com/GeoDB-Limited/odin-core/x/telemetry/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/spf13/cobra"
	"strconv"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	coinswapCmd := &cobra.Command{
		Use:                        telemetrytypes.ModuleName,
		Short:                      "Querying commands for the telemetry module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	coinswapCmd.AddCommand(
		GetQueryCmdTopBalances(),
		GetQueryCmdExtendedValidators(),
	)
	return coinswapCmd
}

// GetQueryCmdTopBalances implements the query parameters command.
// TODO: refactor limit, offset and desc params to unnecessary flags (everywhere)
func GetQueryCmdTopBalances() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "top-balances [denom] [limit] [offset] [desc]",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			denom := args[0]

			var limit uint64 = query.DefaultLimit
			var offset uint64
			if len(args) >= 3 {
				limit, err = strconv.ParseUint(args[1], 10, 64)
				if err != nil {
					return err
				}
				offset, err = strconv.ParseUint(args[2], 10, 64)
				if err != nil {
					return err
				}
			}
			desc := false
			if len(args) == 4 {
				desc, err = strconv.ParseBool(args[3])
				if err != nil {
					return err
				}
			}

			queryClient := telemetrytypes.NewQueryClient(clientCtx)
			res, err := queryClient.TopBalances(cmd.Context(), &telemetrytypes.QueryTopBalancesRequest{
				Denom: denom,
				Pagination: &query.PageRequest{
					Offset: offset,
					Limit:  limit,
				},
				Desc: desc,
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

func GetQueryCmdExtendedValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "extended-validators [limit] [offset] [status]",
		Args: cobra.MaximumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			var limit uint64 = query.DefaultLimit
			var offset uint64
			var status string

			if len(args) > 0 {
				limit, err = strconv.ParseUint(args[1], 10, 64)
				if err != nil {
					return err
				}
			}
			if len(args) > 1 {
				offset, err = strconv.ParseUint(args[2], 10, 64)
				if err != nil {
					return err
				}
			}
			if len(args) > 2 {
				if len(args) > 0 {
					status = args[0]
				}
			}

			queryClient := telemetrytypes.NewQueryClient(clientCtx)
			res, err := queryClient.ExtendedValidators(cmd.Context(), &telemetrytypes.QueryExtendedValidatorsRequest{
				Status: status,
				Pagination: &query.PageRequest{
					Offset: offset,
					Limit:  limit,
				},
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
