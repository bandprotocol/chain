package cli

import (
	// "encoding/json"
	// "fmt"
	// "net/http"

	"context"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	// sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	// clientcmn "github.com/bandprotocol/chain/x/oracle/client/common"
	"github.com/bandprotocol/chain/x/oracle/types"
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
		// GetQueryCmdRequestSearch(storeKey, cdc),
		// GetQueryCmdValidatorStatus(),
		GetQueryCmdReporters(),
		GetQueryActiveValidators(),
		// GetQueryPendingRequests(storeKey, cdc),
		GetQueryRequestPool(),
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
			queryClient := types.NewQueryClient(clientCtx)
			r, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(r)
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
			queryClient := types.NewQueryClient(clientCtx)
			r, err := queryClient.Counts(context.Background(), &types.QueryCountsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(r)
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
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			r, err := queryClient.DataSource(context.Background(), &types.QueryDataSourceRequest{DataSourceId: id})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(r)
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
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			r, err := queryClient.OracleScript(context.Background(), &types.QueryOracleScriptRequest{OracleScriptId: id})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(r)
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

			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			r, err := queryClient.Request(context.Background(), &types.QueryRequestRequest{RequestId: id})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(r)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// // GetQueryCmdRequestSearch implements the search request command.
// func GetQueryCmdRequestSearch(route string, cdc *codec.Codec) *cobra.Command {
// 	return &cobra.Command{
// 		Use:  "request-search [oracle-script-id] [calldata] [ask-count] [min-count]",
// 		Args: cobra.ExactArgs(4),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			cliCtx := context.NewCLIContext().WithCodec(cdc)
// 			bz, _, err := clientcmn.QuerySearchLatestRequest(route, cliCtx, args[0], args[1], args[2], args[3])
// 			if err != nil {
// 				return err
// 			}
// 			return printOutput(cliCtx, cdc, bz, &types.QueryRequestResult{})
// 		},
// 	}
// }

// // GetQueryCmdValidatorStatus implements the query reporter list of validator command.
// func GetQueryCmdValidatorStatus() *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:  "validator [validator]",
// 		Args: cobra.ExactArgs(1),
// 		RunE: func(cmd *cobra.Command, args []string) error {

// 			cliCtx := context.NewCLIContext().WithCodec(cdc)
// 			bz, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", route, types.QueryValidatorStatus, args[0]))
// 			if err != nil {
// 				return err
// 			}
// 			return printOutput(cliCtx, cdc, bz, &types.ValidatorStatus{})
// 		},
// 	}
// 	flags.AddQueryFlagsToCmd(cmd)

// 	return cmd
// }

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
			queryClient := types.NewQueryClient(clientCtx)
			r, err := queryClient.Reporters(context.Background(), &types.QueryReportersRequest{ValidatorAddress: args[1]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(r)
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
			queryClient := types.NewQueryClient(clientCtx)
			r, err := queryClient.ActiveValidators(context.Background(), &types.QueryActiveValidatorsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(r)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// // GetQueryPendingRequests implements the query pending requests command.
// func GetQueryPendingRequests(route string, cdc *codec.Codec) *cobra.Command {
// 	return &cobra.Command{
// 		Use:  "pending-requests [validator]",
// 		Args: cobra.MaximumNArgs(1),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			cliCtx := context.NewCLIContext().WithCodec(cdc)

// 			path := fmt.Sprintf("custom/%s/%s", route, types.QueryPendingRequests)
// 			if len(args) == 1 {
// 				path += "/" + args[0]
// 			}

// 			bz, _, err := cliCtx.Query(path)
// 			if err != nil {
// 				return err
// 			}

// 			return printOutput(cliCtx, cdc, bz, &[]types.RequestID{})
// 		},
// 	}
// }

// GetQueryRequestPool implements the query request pool command.
func GetQueryRequestPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "request-pool [request-key] [port-id] [channel-id]",
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			r, err := queryClient.RequestPool(
				context.Background(),
				&types.QueryRequestPoolRequest{
					RequestKey: args[0],
					PortId:     args[1],
					ChannelId:  args[2],
				},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(r)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
