package cli

import (
	"github.com/GeoDB-Limited/odin-core/x/mint/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"
)

const (
	flagReceiver = "receiver"
	flagAmount   = "amount"
)

// NewTxCmd returns a root CLI command handler for all x/mint transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Mint transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(NewCmdWithdrawCoinsToAccFromTreasury())

	return txCmd
}

// NewCmdWithdrawCoinsToAccFromTreasury implements minting transaction command.
func NewCmdWithdrawCoinsToAccFromTreasury() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-coins (--receiver [receiver]) (--amount [amount])",
		Short: "Withdraw some coins for account",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			receiverStr, err := cmd.Flags().GetString(flagReceiver)
			if err != nil {
				return sdkerrors.Wrapf(err, "flag: %s", flagReceiver)
			}
			receiver, err := sdk.AccAddressFromBech32(receiverStr)
			if err != nil {
				return sdkerrors.Wrapf(err, "receiver: %s", receiverStr)
			}

			amountStr, err := cmd.Flags().GetString(flagAmount)
			if err != nil {
				return sdkerrors.Wrapf(err, "flag: %s", flagAmount)
			}
			amount, err := sdk.ParseCoinsNormalized(amountStr)
			if err != nil {
				return sdkerrors.Wrapf(err, "amount: %s", amountStr)
			}

			msg := types.NewMsgWithdrawCoinsToAccFromTreasury(amount, receiver, clientCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return sdkerrors.Wrapf(err, "amount: %s receiver: %s", amount, receiverStr)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
