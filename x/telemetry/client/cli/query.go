package cli

import (
	telemetrytypes "github.com/GeoDB-Limited/odin-core/x/telemetry/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/spf13/cobra"
	"strconv"
	"time"
)

const (
	DateFormat = "2006-01-02"
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
		GetQueryCmdAvgBlockSize(),
		GetQueryCmdAvgBlockTime(),
		GetQueryCmdAvgTxFee(),
		GetQueryCmdTxVolume(),
		GetQueryCmdValidatorsBlocks(),
	)
	return coinswapCmd
}

// GetQueryCmdTopBalances implements the query parameters command.
func GetQueryCmdTopBalances() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "top-balances [denom] [limit] [offset] [desc]",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			page, desc, err := ParsePagination(args[1], args[2], args[3])
			if err != nil {
				return sdkerrors.Wrap(err, "failed to parse pagination")
			}

			queryClient := telemetrytypes.NewQueryClient(clientCtx)
			res, err := queryClient.TopBalances(cmd.Context(), &telemetrytypes.QueryTopBalancesRequest{
				Denom:      args[0],
				Pagination: page,
				Desc:       desc,
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

			page, _, err := ParsePagination(args[1], args[2], "")
			if err != nil {
				return sdkerrors.Wrap(err, "failed to parse pagination")
			}

			queryClient := telemetrytypes.NewQueryClient(clientCtx)
			res, err := queryClient.ExtendedValidators(cmd.Context(), &telemetrytypes.QueryExtendedValidatorsRequest{
				Status:     args[0],
				Pagination: page,
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

// GetQueryCmdAvgBlockSize implements the query parameters command.
func GetQueryCmdAvgBlockSize() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "avg-block-size [start-date] [end-date]",
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			startDate, endDate, err := ParseDateInterval(args[0], args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "failed to parse date interval")
			}

			queryClient := telemetrytypes.NewQueryClient(clientCtx)
			res, err := queryClient.AvgBlockSize(cmd.Context(), &telemetrytypes.QueryAvgBlockSizeRequest{
				StartDate: startDate,
				EndDate:   endDate,
			})
			if err != nil {
				return sdkerrors.Wrap(err, "failed to query average block size")
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetQueryCmdAvgBlockTime implements the query parameters command.
func
GetQueryCmdAvgBlockTime() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "avg-block-time [start-date] [end-date]",
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			startDate, endDate, err := ParseDateInterval(args[0], args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "failed to parse date interval")
			}

			queryClient := telemetrytypes.NewQueryClient(clientCtx)
			res, err := queryClient.AvgBlockTime(cmd.Context(), &telemetrytypes.QueryAvgBlockTimeRequest{
				StartDate: startDate,
				EndDate:   endDate,
			})
			if err != nil {
				return sdkerrors.Wrap(err, "failed to query average block time")
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetQueryCmdAvgTxFee implements the query parameters command.
func
GetQueryCmdAvgTxFee() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "avg-tx-fee [start-date] [end-date]",
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			startDate, endDate, err := ParseDateInterval(args[0], args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "failed to parse date interval")
			}

			queryClient := telemetrytypes.NewQueryClient(clientCtx)
			res, err := queryClient.AvgTxFee(cmd.Context(), &telemetrytypes.QueryAvgTxFeeRequest{
				StartDate: startDate,
				EndDate:   endDate,
			})
			if err != nil {
				return sdkerrors.Wrap(err, "failed to query average tx fee")
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetQueryCmdTxVolume implements the query parameters command.
func
GetQueryCmdTxVolume() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "tx-volume [start-date] [end-date]",
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			startDate, endDate, err := ParseDateInterval(args[0], args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "failed to parse date interval")
			}

			queryClient := telemetrytypes.NewQueryClient(clientCtx)
			res, err := queryClient.TxVolume(cmd.Context(), &telemetrytypes.QueryTxVolumeRequest{
				StartDate: startDate,
				EndDate:   endDate,
			})
			if err != nil {
				return sdkerrors.Wrap(err, "failed to query tx volume")
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetQueryCmdValidatorsBlocks implements the query parameters command.
func
GetQueryCmdValidatorsBlocks() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "validators-blocks [start-date] [end-date] [limit] [offset] [desc]",
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return sdkerrors.Wrap(err, "failed to get client context")
			}

			startDate, endDate, err := ParseDateInterval(args[0], args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "failed to parse date interval")
			}

			page, desc, err := ParsePagination(args[1], args[2], args[3])
			if err != nil {
				return sdkerrors.Wrap(err, "failed to parse pagination")
			}

			queryClient := telemetrytypes.NewQueryClient(clientCtx)
			res, err := queryClient.ValidatorsBlocks(cmd.Context(), &telemetrytypes.QueryValidatorsBlocksRequest{
				StartDate:  startDate,
				EndDate:    endDate,
				Pagination: page,
				Desc:       desc,
			})
			if err != nil {
				return sdkerrors.Wrap(err, "failed to query validators blocks")
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func ParseDateInterval(startDateArg, endDateArg string) (time.Time, time.Time, error) {
	startDate, err := time.Parse(DateFormat, startDateArg)
	if err != nil {
		return time.Time{}, time.Time{}, sdkerrors.Wrap(err, "failed to parse start date")
	}
	endDate, err := time.Parse(DateFormat, endDateArg)
	if err != nil {
		return time.Time{}, time.Time{}, sdkerrors.Wrap(err, "failed to parse end date")
	}
	return startDate, endDate, err
}

func ParsePagination(limitArg, offsetArg, descArg string) (*query.PageRequest, bool, error) {
	page := &query.PageRequest{
		Offset: 0,
		Limit:  query.DefaultLimit,
	}

	if limitArg != "" {
		limit, err := strconv.ParseUint(limitArg, 10, 64)
		if err != nil {
			return nil, false, sdkerrors.Wrap(err, "failed to parse pagination limit")
		}
		page.Limit = limit
	}
	if offsetArg != "" {
		offset, err := strconv.ParseUint(offsetArg, 10, 64)
		if err != nil {
			return nil, false, sdkerrors.Wrap(err, "failed to parse pagination offset")
		}
		page.Offset = offset
	}
	if descArg != "" {
		desc, err := strconv.ParseBool(descArg)
		if err != nil {
			return nil, false, sdkerrors.Wrap(err, "failed to parse pagination desc")
		}
		return page, desc, nil
	}

	return page, false, nil
}
