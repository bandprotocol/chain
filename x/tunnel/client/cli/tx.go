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

	txCmd.AddCommand(GetTxCmdCreateTSSTunnel())
	txCmd.AddCommand(GetTxCmdCreateIBCTunnel())
	txCmd.AddCommand(GetTxCmdActivate())
	txCmd.AddCommand(GetTxCmdDeactivate())
	txCmd.AddCommand(GetTxCmdTriggerTunnel())

	return txCmd
}

func GetTxCmdCreateTSSTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-tss-tunnel [encoder] [destination-chain-id] [destination-contract-address] [deposit] [signalDeviations-json-file]",
		Short: "Create a new TSS tunnel",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			encoder, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(args[3])
			if err != nil {
				return err
			}

			signalDeviations, err := parseSignalDeviations(args[4])
			if err != nil {
				return err
			}

			interval, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg, err := types.NewMsgCreateTSSTunnel(
				signalDeviations.ToSignalDeviations(),
				interval,
				types.Encoder(encoder),
				args[1],
				args[2],
				deposit,
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

func GetTxCmdCreateIBCTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-ibc-tunnel [channel-id] [deposit] [encoder] [interval] [signalInfos-json-file]",
		Short: "Create a new IBC tunnel",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			encoder, err := strconv.ParseInt(args[2], 10, 32)
			if err != nil {
				return err
			}

			interval, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return err
			}

			signalInfos, err := parseSignalDeviations(args[4])
			if err != nil {
				return err
			}

			msg, err := types.NewMsgCreateIBCTunnel(
				signalInfos.ToSignalDeviations(),
				interval,
				args[0],
				types.Encoder(encoder),
				deposit,
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
