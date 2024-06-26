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
		GetQueryCmdValidatorPrice(),
		GetQueryCmdSignalTotalPowers(),
		GetQueryCmdParams(),
		GetQueryCmdDelegatorSignal(),
		GetQueryCmdSupportedFeeds(),
		GetQueryCmdIsFeeder(),
	)

	return queryCmd
}

// GetQueryCmdDelegatorSignal implements the query delegator signal command.
func GetQueryCmdDelegatorSignal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegator-signal [delegator-addr]",
		Short: "Shows delegator's currently active signal",
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

// GetQueryCmdPrices implements the query prices command.
func GetQueryCmdPrices() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prices",
		Short: "Shows the latest price of all signal ids",
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

// GetQueryCmdPrice implements the query price command.
func GetQueryCmdPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price [signal_id]",
		Short: "Shows the latest price of a signal id",
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

// GetQueryCmdSupportedFeeds implements the query supported feeds command.
func GetQueryCmdSupportedFeeds() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supported-feeds",
		Short: "Shows all currently supported feeds",
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

// GetQueryCmdValidatorPrices implements the query validator prices command.
func GetQueryCmdValidatorPrices() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator-prices [validator]",
		Short: "Shows all prices of the validator",
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

// GetQueryCmdValidatorPrice implements the query validator price command.
func GetQueryCmdValidatorPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price-validator [signal_id] [validator]",
		Short: "Shows the price of validator of the signal id",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.ValidatorPrice(context.Background(), &types.QueryValidatorPriceRequest{
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

// GetQueryCmdSignalTotalPowers implements the query signal-total-powers command.
func GetQueryCmdSignalTotalPowers() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signal-total-powers",
		Short: "Shows all information of all signals and its total power",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.SignalTotalPowers(
				context.Background(),
				&types.QuerySignalTotalPowersRequest{Pagination: pageReq},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "signal-total-powers")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryCmdPriceService implements the query price service command.
func GetQueryCmdPriceService() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price-service",
		Short: "Shows information of price service",
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

// GetQueryCmdParams implements the query params command.
func GetQueryCmdParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Shows the parameters of the module",
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

// GetQueryCmdIsFeeder implements the query if an address is a feeder command.
func GetQueryCmdIsFeeder() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "is-feeder [validator-address] [feeder-address]",
		Short: "Checks if the given address is a feeder for the validator",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.IsFeeder(context.Background(), &types.QueryIsFeederRequest{
				ValidatorAddress: args[0],
				FeederAddress:    args[1],
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
