package cli

import (
	"fmt"
	coinswaptypes "github.com/GeoDB-Limited/odin-core/x/coinswap/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
	"strings"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	oracleCmd := &cobra.Command{
		Use:                        coinswaptypes.ModuleName,
		Short:                      "coinswap transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	oracleCmd.AddCommand(
		GetCmdExchange(),
	)

	return oracleCmd
}

// GetCmdExchange implements the request command handler.
func GetCmdExchange() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exchange [from-denom] [to-denom] [amount]",
		Short: "Exchange the specific amount of one token to another",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Performs exchange of coins denominations according to current rate.
Example:
$ %s tx coinswap exchange geo loki 10loki --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			err = sdk.ValidateDenom(args[0])
			if err != nil {
				return err
			}

			err = sdk.ValidateDenom(args[1])
			if err != nil {
				return err
			}

			amt, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			msg := coinswaptypes.NewMsgExchange(
				args[0],
				args[1],
				amt,
				clientCtx.GetFromAddress(),
			)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}
