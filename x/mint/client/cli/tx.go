package cli

import (
	minttypes "github.com/GeoDB-Limited/odin-core/x/mint/types"
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
		Use:                        minttypes.ModuleName,
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
		Use:   "withdraw-coins [receiver] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Withdraw some coins for account",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			receiver, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return sdkerrors.Wrapf(err, "receiver: %s", args[0])
			}

			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return sdkerrors.Wrapf(err, "amount: %s", args[1])
			}

			msg := minttypes.NewMsgWithdrawCoinsToAccFromTreasury(amount, receiver, clientCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return sdkerrors.Wrapf(err, "amount: %s receiver: %s", amount, receiver)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
