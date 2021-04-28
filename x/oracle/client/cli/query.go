package cli

import (
	"fmt"
	oracleclientcommon "github.com/GeoDB-Limited/odin-core/x/oracle/client/common"
	"github.com/GeoDB-Limited/odin-core/x/oracle/types"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	oracleCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the oracle module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	oracleCmd.AddCommand(
		GetQueryCmdParams(),
		GetQueryCmdCounts(),
		GetQueryCmdDataSource(),
		GetQueryCmdOracleScript(),
		GetQueryCmdRequest(),
		GetQueryCmdRequestSearch(),
		GetQueryCmdValidatorStatus(),
		GetQueryCmdReporters(),
		GetQueryActiveValidators(),
		GetCmdQueryDataProvidersPool(),
		GetCmdQueryRequestPrice(),
		GetQueryCmdData(),
	)
	return oracleCmd
}

// GetQueryCmdParams implements the query parameters command.
func GetQueryCmdParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "params",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := oracletypes.NewQueryClient(clientCtx)
			res, err := queryClient.Params(cmd.Context(), &oracletypes.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetQueryCmdCounts implements the query counts command.
func GetQueryCmdCounts() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "counts",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := oracletypes.NewQueryClient(clientCtx)
			res, err := queryClient.Counts(cmd.Context(), &oracletypes.QueryCountsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetQueryCmdDataSource implements the query data source command.
func GetQueryCmdDataSource() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "data-source [id]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			dsId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := oracletypes.NewQueryClient(clientCtx)
			res, err := queryClient.DataSource(cmd.Context(), &oracletypes.QueryDataSourceRequest{
				DataSourceId: dsId,
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

// GetQueryCmdDataSources implements the query data sources command.
func GetQueryCmdDataSources() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "data-sources [page] [limit]",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			page, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			limit, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			queryClient := oracletypes.NewQueryClient(clientCtx)
			res, err := queryClient.DataSources(cmd.Context(), &oracletypes.QueryDataSourcesRequest{
				Page:  page,
				Limit: limit,
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

// GetQueryCmdOracleScript implements the query oracle script command.
func GetQueryCmdOracleScript() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "oracle-script [id]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			osId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := oracletypes.NewQueryClient(clientCtx)
			res, err := queryClient.OracleScript(cmd.Context(), &oracletypes.QueryOracleScriptRequest{
				OracleScriptId: osId,
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

// GetQueryCmdRequest implements the query request command.
func GetQueryCmdRequest() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "request [id]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			rId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := oracletypes.NewQueryClient(clientCtx)
			res, err := queryClient.Request(cmd.Context(), &oracletypes.QueryRequestRequest{
				RequestId: rId,
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

// GetQueryCmdRequestSearch implements the search request command.
func GetQueryCmdRequestSearch() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "request-search (-s [oracle-script-id]) (-c [calldata]) (-a [ask-count]) (-m [min-count])",
		Args: cobra.RangeArgs(1, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			oid, err := cmd.Flags().GetInt64(flagOracleScriptID)
			if err != nil {
				return err
			}

			callData, err := cmd.Flags().GetBytesHex(flagCalldata)
			if err != nil {
				return err
			}

			askCount, err := cmd.Flags().GetInt64(flagAskCount)
			if err != nil {
				return err
			}

			minCount, err := cmd.Flags().GetInt64(flagMinCount)
			if err != nil {
				return err
			}

			res, _, err := oracleclientcommon.QuerySearchLatestRequest(oracletypes.QuerierRoute, clientCtx, &oracletypes.QueryRequestSearchRequest{
				OracleScriptId: oid,
				Calldata:       callData,
				AskCount:       askCount,
				MinCount:       minCount,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	cmd.Flags().BytesHexP(flagCalldata, "c", nil, "Calldata used in calling the oracle script")
	cmd.Flags().Int64P(flagOracleScriptID, "s", 0, "oracle script id used in request")
	cmd.Flags().Int64P(flagMinCount, "m", 0, "min count used in request")
	cmd.Flags().Int64P(flagAskCount, "a", 0, "ask count used in request")

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetQueryCmdValidatorStatus implements the query reporter list of validator command.
func GetQueryCmdValidatorStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "validator [validator]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := oracletypes.NewQueryClient(clientCtx)
			res, err := queryClient.Validator(cmd.Context(), &oracletypes.QueryValidatorRequest{
				ValidatorAddress: args[0],
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

// GetQueryCmdReporters implements the query reporter list of validator command.
func GetQueryCmdReporters() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "reporters [validator]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := oracletypes.NewQueryClient(clientCtx)
			res, err := queryClient.Reporters(cmd.Context(), &oracletypes.QueryReportersRequest{
				ValidatorAddress: args[0],
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

// GetQueryActiveValidators implements the query active validators command.
func GetQueryActiveValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "active-validators",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := oracletypes.NewQueryClient(clientCtx)
			res, err := queryClient.ActiveValidators(cmd.Context(), &oracletypes.QueryActiveValidatorsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryDataProvidersPool returns the command for fetching community pool info
func GetCmdQueryDataProvidersPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data-providers-pool",
		Args:  cobra.NoArgs,
		Short: "Query the amount of coins in the data providers pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all coins in the data providers pool.

Example:
$ %s query oracle data-providers-pool
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := oracletypes.NewQueryClient(clientCtx)
			res, err := queryClient.DataProvidersPool(cmd.Context(), &oracletypes.QueryDataProvidersPoolRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetCmdQueryRequestPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request-price",
		Args:  cobra.NoArgs,
		Short: "queries the latest price on standard price reference oracle",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			symbol := args[0]
			askCount, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}
			minCount, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return err
			}

			queryClient := oracletypes.NewQueryClient(clientCtx)
			res, err := queryClient.RequestPrice(cmd.Context(), &oracletypes.QueryRequestPriceRequest{
				Symbol:   symbol,
				AskCount: askCount,
				MinCount: minCount,
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

func GetQueryCmdData() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "data [data-hash]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := oracletypes.NewQueryClient(clientCtx)
			res, err := queryClient.Data(cmd.Context(), &oracletypes.QueryDataRequest{
				DataHash: args[0],
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
