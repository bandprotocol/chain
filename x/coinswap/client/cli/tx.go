package cli

import (
	"bufio"
	"fmt"
	"github.com/GeoDB-Limited/odincore/chain/x/coinswap/types"
	commontypes "github.com/GeoDB-Limited/odincore/chain/x/common/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"
	"strings"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	oracleCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "coinswap transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	oracleCmd.AddCommand(flags.PostCommands(
		GetCmdExchange(cdc),
	)...)

	return oracleCmd
}

// GetCmdExchange implements the request command handler.
func GetCmdExchange(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exchange [from-denom] [to-denom] [amount]",
		Short: "Exchange the specific amount of one token to another",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Performs exchange of coins denominations according to current rate.
Example:
$ %s tx coinswap exchange geo loki 10loki --from mykey
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			fromDenom, err := commontypes.ParseDenom(args[0])
			if err != nil {
				return err
			}

			toDenom, err := commontypes.ParseDenom(args[1])
			if err != nil {
				return err
			}

			amt, err := sdk.ParseCoin(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgExchange(
				fromDenom,
				toDenom,
				amt,
				cliCtx.GetFromAddress(),
			)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}
