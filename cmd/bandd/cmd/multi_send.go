package cmd

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/cobra"
)

// MultiSendTxCmd creates a multi-send tx and signs it with the given key.
func MultiSendTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multi-send [amount_per_account] [to_address1] [to_address2] ....",
		Short: "Send token to multiple accounts",
		Long: "Send equal amount of token to multiple accounts",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			sender := sdk.AccAddress(clientCtx.GetFromAddress())

			// Parse the coins we are trying to send
			coins, err := sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return err
			}
			accounts := args[1:]
			inputCoins := sdk.NewCoins()
			outputs := make([]banktypes.Output, 0, len(accounts))
			for _, acc := range accounts {
				to, err := sdk.AccAddressFromBech32(acc)
				if err != nil {
					return err
				}
				outputs = append(outputs, banktypes.NewOutput(to, coins))
				inputCoins = inputCoins.Add(coins...)
			}
			msg := banktypes.NewMsgMultiSend(
				[]banktypes.Input{banktypes.NewInput(sender, inputCoins)},
				outputs,
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)

		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
