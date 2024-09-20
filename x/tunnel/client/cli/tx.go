package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// GetTxCmd returns a root CLI command handler for all x/tunnel transaction commands.
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "tunnel transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(GetTxCmdCreateTunnels())
	txCmd.AddCommand(GetTxCmdEditTunnel())
	txCmd.AddCommand(GetTxCmdActivate())
	txCmd.AddCommand(GetTxCmdDeactivate())
	txCmd.AddCommand(GetTxCmdTriggerTunnel())
	txCmd.AddCommand(GetTxCmdDepositTunnel())
	txCmd.AddCommand(GetTxCmdWithdrawTunnel())

	return txCmd
}

func GetTxCmdCreateTunnels() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                "create",
		Short:              "Create a new tunnel",
		DisableFlagParsing: true,
		RunE:               client.ValidateCmd,
	}

	txCmd.AddCommand(GetTxCmdCreateTSSTunnel())

	flags.AddTxFlagsToCmd(txCmd)

	return txCmd
}

func GetTxCmdCreateTSSTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tss [destination-chain-id] [destination-contract-address] [encoder] [initial-deposit] [interval] [signalDeviations-json-file]",
		Short: "Create a new TSS tunnel",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			encoder, err := strconv.ParseInt(args[2], 10, 32)
			if err != nil {
				return err
			}

			initialDeposit, err := sdk.ParseCoinsNormalized(args[3])
			if err != nil {
				return err
			}

			interval, err := strconv.ParseUint(args[4], 10, 64)
			if err != nil {
				return err
			}

			signalDeviations, err := parseSignalDeviations(args[5])
			if err != nil {
				return err
			}

			msg, err := types.NewMsgCreateTSSTunnel(
				signalDeviations.ToSignalDeviations(),
				interval,
				args[0],
				args[1],
				types.Encoder(encoder),
				initialDeposit,
				clientCtx.GetFromAddress(),
			)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func GetTxCmdEditTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit [tunnel-id] [interval] [signalDeviations-json-file] ",
		Short: "Edit an existing tunnel",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			interval, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			signalDeviations, err := parseSignalDeviations(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgEditTunnel(
				id,
				signalDeviations.ToSignalDeviations(),
				interval,
				clientCtx.GetFromAddress().String(),
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func GetTxCmdActivate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate [tunnel-id]",
		Short: "Activate a tunnel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgActivate(id, clientCtx.GetFromAddress().String())
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func GetTxCmdDeactivate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate [tunnel-id]",
		Short: "Deactivate a tunnel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeactivate(id, clientCtx.GetFromAddress().String())
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func GetTxCmdTriggerTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger [tunnel-id]",
		Short: "Manual trigger a tunnel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			tunnelID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgTriggerTunnel(tunnelID, clientCtx.GetFromAddress().String())
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func GetTxCmdDepositTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit-tunnel [tunnel-id] [amount]",
		Short: "Deposit to a tunnel",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgDepositTunnel(id, amount, clientCtx.GetFromAddress().String())
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func GetTxCmdWithdrawTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw [tunnel-id] [amount]",
		Short: "Withdraw deposit from a tunnel",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdrawTunnel(id, amount, clientCtx.GetFromAddress().String())
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
