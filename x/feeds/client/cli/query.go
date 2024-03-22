package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		GetQueryCmdPrices(),
		GetQueryCmdPrice(),
		GetQueryCmdValidatorPrices(),
		GetQueryCmdPriceValidator(),
		GetQueryCmdFeeds(),
		GetQueryCmdParams(),
		GetQueryCmdDelegatorSignal(),
		GetQueryCmdSupportedFeeds(),
	)

	return queryCmd
}

func GetQueryCmdDelegatorSignal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegator-signal [delegator-addr]",
		Short: "shows delegator's currently active signal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.DelegatorSignals(
				context.Background(),
				&types.QueryDelegatorSignalsRequest{Delegator: args[0]},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetQueryCmdPrices() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prices",
		Short: "shows the latest price of all signal ids",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.Prices(context.Background(), &types.QueryPricesRequest{Pagination: pageReq})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "prices")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetQueryCmdPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price [signal_id]",
		Short: "shows the latest price of a signal id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Price(context.Background(), &types.QueryPriceRequest{
				SignalId: args[0],
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

func GetQueryCmdSupportedFeeds() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supported-feeds",
		Short: "shows all currently supported feeds",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.SupportedFeeds(context.Background(), &types.QuerySupportedFeedsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetQueryCmdValidatorPrices() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator-prices [validator]",
		Short: "shows all prices of the validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.ValidatorPrices(context.Background(), &types.QueryValidatorPricesRequest{
				Validator: args[0],
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

func GetQueryCmdPriceValidator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price-validator [signal_id] [validator]",
		Short: "shows the price of validator of the signal id",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.PriceValidator(context.Background(), &types.QueryPriceValidatorRequest{
				SignalId:  args[0],
				Validator: args[1],
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

func GetQueryCmdFeeds() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feeds",
		Short: "shows all information of all feeds",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.Feeds(context.Background(), &types.QueryFeedsRequest{Pagination: pageReq})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "feeds")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetQueryCmdPriceService() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price-service",
		Short: "shows information of price service",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.PriceService(context.Background(), &types.QueryPriceServiceRequest{})
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
		Short: "shows the parameters of the module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
