package cli

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
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

	txCmd.AddCommand(
		GetTxCmdCreateTunnel(),
		GetTxCmdUpdateRoute(),
		GetTxCmdUpdateSignalsAndInterval(),
		GetTxCmdWithdrawFeePayerFunds(),
		GetTxCmdActivate(),
		GetTxCmdDeactivate(),
		GetTxCmdTriggerTunnel(),
		GetTxCmdDepositToTunnel(),
		GetTxCmdWithdrawFromTunnel(),
	)

	return txCmd
}

func GetTxCmdCreateTunnel() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                "create-tunnel",
		Short:              "Create a new tunnel",
		DisableFlagParsing: true,
		RunE:               client.ValidateCmd,
	}

	// add create tunnel subcommands
	txCmd.AddCommand(
		GetTxCmdCreateTSSTunnel(),
		GetTxCmdCreateIBCTunnel(),
		GetTxCmdCreateRouterTunnel(),
	)

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

			destChainID := args[0]
			destContractAddr := args[1]

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
				destChainID,
				destContractAddr,
				feedstypes.Encoder(encoder),
				initialDeposit,
				clientCtx.GetFromAddress().String(),
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

func GetTxCmdCreateIBCTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ibc [initial-deposit] [interval] [signalInfos-json-file]",
		Short: "Create a new IBC tunnel",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			initialDeposit, err := sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return err
			}

			interval, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			signalInfos, err := parseSignalDeviations(args[2])
			if err != nil {
				return err
			}

			msg, err := types.NewMsgCreateIBCTunnel(
				signalInfos.ToSignalDeviations(),
				interval,
				initialDeposit,
				clientCtx.GetFromAddress().String(),
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

func GetTxCmdCreateRouterTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "router [fund] [bridge-contract-address] [dest-chain-id] [dest-contract-address] [dest-gas-limit] [dest-gas-price] [initial-deposit] [interval] [signalDeviations-json-file]",
		Short: "Create a new router tunnel",
		Args:  cobra.ExactArgs(9),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			fund, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			bridgeContractAddr := args[1]
			destChainID := args[2]
			destContractAddr := args[3]

			destGasLimit, err := strconv.ParseUint(args[4], 10, 64)
			if err != nil {
				return err
			}

			destGasPrice, err := strconv.ParseUint(args[5], 10, 64)
			if err != nil {
				return err
			}

			initialDeposit, err := sdk.ParseCoinsNormalized(args[6])
			if err != nil {
				return err
			}

			interval, err := strconv.ParseUint(args[7], 10, 64)
			if err != nil {
				return err
			}

			signalDeviations, err := parseSignalDeviations(args[8])
			if err != nil {
				return err
			}

			msg, err := types.NewMsgCreateRouterTunnel(
				signalDeviations.ToSignalDeviations(),
				interval,
				fund,
				bridgeContractAddr,
				destChainID,
				destContractAddr,
				destGasLimit,
				destGasPrice,
				initialDeposit,
				clientCtx.GetFromAddress().String(),
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

func GetTxCmdUpdateRoute() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                "update-route",
		Short:              "Update route information",
		DisableFlagParsing: true,
		RunE:               client.ValidateCmd,
	}

	// add create tunnel subcommands
	txCmd.AddCommand(
		GetTxCmdUpdateIBCRoute(),
	)

	return txCmd
}

func GetTxCmdUpdateIBCRoute() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ibc [tunnel-id] [channel-id]",
		Short: "Update IBC route of a IBC tunnel",
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

			msg, err := types.NewMsgUpdateIBCRoute(
				id,
				args[1],
				clientCtx.GetFromAddress().String(),
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

func GetTxCmdUpdateSignalsAndInterval() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-and-reset-tunnel [tunnel-id] [interval] [signalDeviations-json-file] ",
		Short: "Update an existing tunnel and reset the latest price interval of the tunnel",
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

			msg := types.NewMsgUpdateSignalsAndInterval(
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

func GetTxCmdWithdrawFeePayerFunds() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-fee-payer-funds [tunnel-id] [amount]",
		Short: "Withdraw fee payer funds from a tunnel to the creator",
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

			msg := types.NewMsgWithdrawFeePayerFunds(id, amount, clientCtx.GetFromAddress().String())
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
		Use:   "trigger-tunnel [tunnel-id]",
		Short: "Trigger a tunnel to generate a new packet",
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

func GetTxCmdDepositToTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit-to-tunnel [tunnel-id] [amount]",
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

			msg := types.NewMsgDepositToTunnel(id, amount, clientCtx.GetFromAddress().String())
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func GetTxCmdWithdrawFromTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-from-tunnel [tunnel-id] [amount]",
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

			msg := types.NewMsgWithdrawFromTunnel(id, amount, clientCtx.GetFromAddress().String())
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
