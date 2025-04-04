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
		GetTxCmdActivateTunnel(),
		GetTxCmdDeactivateTunnel(),
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
		GetTxCmdCreateIBCHookTunnel(),
		GetTxCmdCreateAxelarTunnel(),
		GetTxCmdCreateRouterTunnel(),
	)

	return txCmd
}

func GetTxCmdCreateTSSTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tss [destination-chain-id] [destination-contract-address] [encoder] [initial-deposit] [interval] [signal-deviations-json-file]",
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
		Use:   "ibc [initial-deposit] [interval] [signal-deviations-json-file]",
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

			signalDeviations, err := parseSignalDeviations(args[2])
			if err != nil {
				return err
			}

			msg, err := types.NewMsgCreateIBCTunnel(
				signalDeviations.ToSignalDeviations(),
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

func GetTxCmdCreateIBCHookTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ibc-hook [channel-id] [destination-contract-address] [initial-deposit] [interval] [signal-deviations-json-file]",
		Short: "Create a new IBC hook tunnel",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			channelID := args[0]
			destinationContractAddress := args[1]

			initialDeposit, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return err
			}

			interval, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return err
			}

			signalDeviations, err := parseSignalDeviations(args[4])
			if err != nil {
				return err
			}

			msg, err := types.NewMsgCreateIBCHookTunnel(
				signalDeviations.ToSignalDeviations(),
				interval,
				channelID,
				destinationContractAddress,
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

func GetTxCmdCreateAxelarTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "axelar [destination-chain-id] [destination-address] [axelar-fee] [initial-deposit] [interval] [signal-deviations-json-file]",
		Short: "Create a new Axelar tunnel",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			destinationChainID := args[0]
			destinationContractAddress := args[1]

			axelarFee, err := sdk.ParseCoinNormalized(args[2])
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

			signalInfos, err := parseSignalDeviations(args[5])
			if err != nil {
				return err
			}

			msg, err := types.NewMsgCreateAxelarTunnel(
				signalInfos.ToSignalDeviations(),
				interval,
				destinationChainID,
				destinationContractAddress,
				axelarFee,
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
		Use:   "router [destination-chain-id] [destination-contract-address] [destination-gas-limit] [initial-deposit] [interval] [signal-deviations-json-file]",
		Short: "Create a new Router tunnel",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			destinationChainID := args[0]
			destinationContractAddress := args[1]

			destinationGasLimit, err := strconv.ParseUint(args[2], 10, 64)
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

			msg, err := types.NewMsgCreateRouterTunnel(
				signalDeviations.ToSignalDeviations(),
				interval,
				destinationChainID,
				destinationContractAddress,
				destinationGasLimit,
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
		GetTxCmdUpdateIBCHookRoute(),
		GetTxCmdUpdateAxelarRoute(),
		GetTxCmdUpdateRouterRoute(),
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

func GetTxCmdUpdateIBCHookRoute() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ibc-hook [tunnel-id] [channel-id] [destination-contract-address]",
		Short: "Update IBC route of a IBC tunnel",
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

			channelID := args[1]
			destContractAddr := args[2]

			msg, err := types.NewMsgUpdateIBCHookRoute(
				id,
				channelID,
				destContractAddr,
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

func GetTxCmdUpdateAxelarRoute() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "axelar [tunnel-id] [destination-chain-id] [destination-contract-address] [axelar-fee]",
		Short: "Update Axelar route of a Axelar tunnel",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			destinationChainID := args[1]
			destinationContractAddress := args[2]

			axelarFee, err := sdk.ParseCoinNormalized(args[3])
			if err != nil {
				return err
			}

			msg, err := types.NewMsgUpdateAxelarRoute(
				id,
				destinationChainID,
				destinationContractAddress,
				axelarFee,
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

func GetTxCmdUpdateRouterRoute() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "router [tunnel-id] [destination-chain-id] [destination-contract-address] [destination-gas-limit]",
		Short: "Update Router route of a Router tunnel",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			destinationChainID := args[1]
			destinationContractAddr := args[2]

			destinationGasLimit, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return err
			}

			msg, err := types.NewMsgUpdateRouterRoute(
				id,
				destinationChainID,
				destinationContractAddr,
				destinationGasLimit,
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
		Use:   "update-signals-and-interval [tunnel-id] [interval] [signalDeviations-json-file] ",
		Short: "Update signals and interval of the existing tunnel",
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

func GetTxCmdActivateTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate-tunnel [tunnel-id]",
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

			msg := types.NewMsgActivateTunnel(id, clientCtx.GetFromAddress().String())
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func GetTxCmdDeactivateTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate-tunnel [tunnel-id]",
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

			msg := types.NewMsgDeactivateTunnel(id, clientCtx.GetFromAddress().String())
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
