package cli

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
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
		GetQueryCmdValidatorStatus(),
		GetQueryCmdReporters(),
		GetQueryActiveValidators(),
		GetQueryPendingRequests(),
		GetQueryRequestVerification(),
		GetQueryRequestPool(),
	)
	return oracleCmd
}

// GetQueryCmdParams implements the query parameters command.
func GetQueryCmdParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Get current parameters of Bandchain's oracle module",
		Args:  cobra.NoArgs,
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
		Use:   "counts",
		Short: "Get number of requests, oracle scripts, and data source scripts currently deployed on Bandchain",
		Args:  cobra.NoArgs,
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
		Use:   "data-source [id]",
		Short: "Get summary information of a data source",
		Args:  cobra.ExactArgs(1),
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
		Use:   "oracle-script [id]",
		Short: "Get summary information of an oracle script",
		Args:  cobra.ExactArgs(1),
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
		Use:   "request [id]",
		Short: "Get an oracle request details",
		Args:  cobra.ExactArgs(1),
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

// GetQueryCmdValidatorStatus implements the query of validator status.
func GetQueryCmdValidatorStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator [validator-address]",
		Short: "Get active status of a validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			valAddress, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			r, err := queryClient.Validator(context.Background(), &types.QueryValidatorRequest{ValidatorAddress: valAddress.String()})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(r)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryCmdReporters implements the query reporter list of validator command.
func GetQueryCmdReporters() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reporters [validator-address]",
		Short: "Get list of reporters owned by given validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			r, err := queryClient.Reporters(context.Background(), &types.QueryReportersRequest{ValidatorAddress: args[0]})
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
		Use:   "active-validators",
		Short: "Get number of active validators",
		Args:  cobra.NoArgs,
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

// GetQueryPendingRequests implements the query pending requests command.
func GetQueryPendingRequests() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-requests [validator-address]",
		Short: "Get list of pending request IDs assigned to given validator",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			valAddress, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("unable to parse given validator address: %w", err)
			}

			r, err := queryClient.PendingRequests(context.Background(), &types.QueryPendingRequestsRequest{
				ValidatorAddress: valAddress.String(),
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(r)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryRequestVerification implements the query request verification command.
func GetQueryRequestVerification() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify-request [chain-id] [validator-addr] [request-id] [data-source-external-id] [reporter-pubkey] [reporter-signature-hex]",
		Short: "Verify validity of pending oracle requests",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			requestID, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("unable to parse request ID: %w", err)
			}
			externalID, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return fmt.Errorf("unable to parse external ID: %w", err)
			}

			signature, err := hex.DecodeString(args[5])
			if err != nil {
				return fmt.Errorf("unable to parse signature: %w", err)
			}

			r, err := queryClient.RequestVerification(context.Background(), &types.QueryRequestVerificationRequest{
				ChainId:    args[0],
				Validator:  args[1],
				RequestId:  requestID,
				ExternalId: externalID,
				Reporter:   args[4],
				Signature:  signature,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(r)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryRequestPool implements the query request pool command.
func GetQueryRequestPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request-pool [request-key] [port-id] [channel-id]",
		Short: "Get account information of request pool",
		Args:  cobra.ExactArgs(3),
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
