package cli

import (
	"fmt"
	auctiontypes "github.com/GeoDB-Limited/odin-core/x/auction/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
	"strings"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	auctionCmd := &cobra.Command{
		Use:                        auctiontypes.ModuleName,
		Short:                      "auction transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	auctionCmd.AddCommand(
		GetCmdBuyCoins(),
	)

	return auctionCmd
}

// GetCmdBuyCoins implements the request command handler.
func GetCmdBuyCoins() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buy-coins [from-denom] [to-denom] [amount]",
		Short: "Buy amount of coins for another coins",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Performs exchange of coins denominations according to current rate.
Example:
$ %s tx auction buy-coins minigeo loki 10minigeo --from mykey
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

			msg := auctiontypes.NewMsgBuyCoins(
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

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
