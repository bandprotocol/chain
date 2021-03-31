package cli

import (
	"fmt"
	"github.com/GeoDB-Limited/odincore/chain/x/coinswap/types"
	"github.com/GeoDB-Limited/odincore/chain/x/common/client/cli"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	coinswapCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the coinswap module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	coinswapCmd.AddCommand(flags.GetCommands(
		GetQueryCmdParams(storeKey, cdc),
		GetQueryCmdRate(storeKey, cdc),
	)...)
	return coinswapCmd
}

// GetQueryCmdParams implements the query parameters command.
func GetQueryCmdParams(route string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:  "params",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			bz, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", route, types.QueryParams))
			if err != nil {
				return err
			}
			return cli.PrintOutput(cliCtx, cdc, bz, &types.Params{})
		},
	}
}

func GetQueryCmdRate(route string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:  "rate [from-denom] [to-denom]",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			bz, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", route, types.QueryRate))
			if err != nil {
				return err
			}
			return cli.PrintOutput(cliCtx, cdc, bz, &types.QueryRateResult{})
		},
	}
}
